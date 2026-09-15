package dispatchserver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"myGinServer/api/request"
	istore "myGinServer/internal/store/dispatch"
	mdispatch "myGinServer/models/dispatch"
)

// drainCh 清空全局 channel，避免用例之间互相影响。
func drainCh() {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}

// TestSendEntity_QueueFullReturnsError 校验队列满时立即返回错误，而不是把调用方挂住。
//
// 修复前的写法是 select{case <-ctx.Done(): ...; default: ch <- ...}，
// default 分支里的发送是阻塞发送：channel 满时调用方会永久阻塞，
// 且因为根本没在监听 ctx，客户端断开也无法解除。
func TestSendEntity_QueueFullReturnsError(t *testing.T) {
	drainCh()
	defer drainCh()
	n := &NodeServer{}

	for i := 0; i < cap(ch); i++ {
		if err := n.SendEntity(context.Background(), "run-fill", "select 1"); err != nil {
			t.Fatalf("填充第 %d 条时不应失败: %v", i, err)
		}
	}

	// 用超时保护：若实现退化成阻塞发送，这里会超时失败，而不是一直挂住整个测试进程。
	done := make(chan error, 1)
	go func() {
		done <- n.SendEntity(context.Background(), "run-overflow", "select 1")
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("队列已满时 SendEntity 应返回错误")
		}
		if !strings.Contains(err.Error(), "队列已满") {
			t.Fatalf("错误信息应说明队列已满，实际为: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("队列已满时 SendEntity 发生阻塞（修复前的行为）")
	}
}

// TestSendEntity_CancelledContext 校验已取消的请求被直接拒绝。
//
// 修复前没有 ctx 前置检查：当 ctx 已取消而 channel 恰有空位时，
// select 会在「发送」与「ctx.Done()」之间随机选一个，约有一半概率仍把
// 已取消的请求投递进队列。
func TestSendEntity_CancelledContext(t *testing.T) {
	drainCh()
	defer drainCh()
	n := &NodeServer{}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := n.SendEntity(ctx, "run-cancel", "select 1"); err != context.Canceled {
		t.Fatalf("期望返回 context.Canceled，实际为: %v", err)
	}
	if len(ch) != 0 {
		t.Fatalf("ctx 已取消时不应投递消息，实际队列长度 %d", len(ch))
	}
}

// TestSendEntity_OK 校验正常路径的投递与内容。
func TestSendEntity_OK(t *testing.T) {
	drainCh()
	defer drainCh()
	n := &NodeServer{}

	if err := n.SendEntity(context.Background(), "run-ok", "select 1"); err != nil {
		t.Fatalf("正常投递不应失败: %v", err)
	}
	if len(ch) != 1 {
		t.Fatalf("应投递 1 条消息，实际 %d", len(ch))
	}
	entity := <-ch
	if entity.RunId != "run-ok" || entity.Sql != "select 1" {
		t.Fatalf("投递内容不正确: %+v", entity)
	}
}

// updateCall 记录一次 exec_queue 状态写入。
type updateCall struct {
	runId  string
	resp   string
	status string
}

// fakeStore 只实现本文件用到的几个方法；其余方法由嵌入的接口提供，
// 一旦被意外调用会因 nil 接口而 panic，从而暴露用例假设不成立。
//
// 注意两类方法的 ctx 处理方式不同，这是刻意为之：
//   - ExecQueueExists / SaveExecQueue 忽略 ctx。用例是拿「一开始就已取消」的 ctx
//     来驱动流程的，若这两个方法也照实拒绝，TestRun 会停在最前面，根本走不到
//     需要验证的补偿分支；真实场景里 ctx 是在请求处理中途才取消的。
//   - UpdateExecQueueResp 严格检查 ctx 并如实返回 ctx.Err()，模拟数据库
//     「ctx 已取消则写入立即失败」的行为。这是本用例真正的断言点：
//     若补偿写入沿用了被取消的 ctx，更新就会丢失，用例随即失败。
type fakeStore struct {
	istore.StoreIface
	saved   mdispatch.ExecQueue
	updated []updateCall
}

func (f *fakeStore) ExecQueueExists(context.Context, string, string) (bool, error) {
	return false, nil
}

func (f *fakeStore) SaveExecQueue(_ context.Context, _ string, entity mdispatch.ExecQueue) error {
	f.saved = entity
	return nil
}

func (f *fakeStore) UpdateExecQueueResp(ctx context.Context, _, runId, resp, status string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.updated = append(f.updated, updateCall{runId: runId, resp: resp, status: status})
	return nil
}

// TestTestRun_SendFailureMarksExecQueueFailed 校验投递失败时的补偿写入。
//
// 关键点有两个：
//  1. ctx 已取消时 TestRun 仍必须把补偿写入落库 —— 所以那里必须用
//     context.WithoutCancel 派生 context，否则「因为 ctx 取消而失败」的场景下，
//     补偿写入自身也会被取消，exec_queue 里的 pending 行依旧永远无人处理。
//  2. 写入的 Response 必须是合法 JSON，因为 QueryTestResult 对 finished 状态
//     会直接 json.Unmarshal(Response)。
func TestTestRun_SendFailureMarksExecQueueFailed(t *testing.T) {
	drainCh()
	defer drainCh()
	// 填满 channel，使 SendEntity 必定走到投递失败分支
	for i := 0; i < cap(ch); i++ {
		ch <- TestRunEntity{RunId: "fill", Sql: "select 1"}
	}

	fs := &fakeStore{}
	n := &NodeServer{store: fs}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 模拟客户端已断开

	err := n.TestRun(ctx, request.TestRunSqlReq{
		RunId:  "run-compensate",
		Sql:    base64.StdEncoding.EncodeToString([]byte("select 1")),
		Params: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("投递失败时 TestRun 应返回错误")
	}

	if fs.saved.RunId != "run-compensate" {
		t.Fatalf("应先落一条 exec_queue 记录，实际: %+v", fs.saved)
	}
	if len(fs.updated) != 1 {
		t.Fatalf("ctx 已取消时补偿写入也必须落库，期望 1 次更新，实际 %d 次", len(fs.updated))
	}
	got := fs.updated[0]
	if got.status != string(mdispatch.Failed) {
		t.Fatalf("状态应被置为 failed，实际 %s", got.status)
	}
	if !json.Valid([]byte(got.resp)) {
		t.Fatalf("Response 必须是合法 JSON，实际: %s", got.resp)
	}
	var resp struct {
		Msg string `json:"msg"`
	}
	if err := json.Unmarshal([]byte(got.resp), &resp); err != nil {
		t.Fatalf("Response 反序列化失败: %v", err)
	}
	if !strings.Contains(resp.Msg, "投递执行队列失败") {
		t.Fatalf("应写明失败原因，实际 msg: %s", resp.Msg)
	}
}
