package delayqueue

import (
	"context"
	"strings"
	"sync"
	"testing"
)

// TestCallHandler_PanicRecovered 校验 handler 内的 panic 被收敛成 error 返回。
//
// 这里的 key point 是「不能打挂测试进程」：修复前 processOne 直接调用 handler，
// worker goroutine 上方没有 recover，一次 panic 就会终止整个进程。
// 若本用例因进程崩溃而失败，说明 recover 未生效。
func TestCallHandler_PanicRecovered(t *testing.T) {
	c := NewConsumer(nil)

	err := c.callHandler(context.Background(), &DelayMessage{}, func(context.Context, *DelayMessage) error {
		panic("boom")
	})

	if err == nil {
		t.Fatal("handler 内 panic 未被捕获，期望返回 error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("错误信息应包含 panic 内容，实际为: %v", err)
	}
}

// TestCallHandler_ErrorAndSuccessPassthrough 校验正常路径不被 recover 逻辑干扰。
func TestCallHandler_ErrorAndSuccessPassthrough(t *testing.T) {
	c := NewConsumer(nil)

	wantErr := context.DeadlineExceeded
	if err := c.callHandler(context.Background(), &DelayMessage{},
		func(context.Context, *DelayMessage) error { return wantErr },
	); err != wantErr {
		t.Fatalf("handler 返回的 error 应原样透传，期望 %v，实际 %v", wantErr, err)
	}

	if err := c.callHandler(context.Background(), &DelayMessage{},
		func(context.Context, *DelayMessage) error { return nil },
	); err != nil {
		t.Fatalf("handler 成功时应返回 nil，实际 %v", err)
	}
}

// TestLookupHandler_ConcurrentWithRegister 校验 handlers 的读写并发安全。
// 修复前 Register 与 lookupHandler 均直接裸访问 map，配合 -race 可复现数据竞争。
func TestLookupHandler_ConcurrentWithRegister(t *testing.T) {
	c := NewConsumer(nil)
	const topic = "node_instance"
	noop := func(context.Context, *DelayMessage) error { return nil }

	var wg sync.WaitGroup
	// 并发写
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.Register(topic, noop)
			}
		}()
	}
	// 并发读
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				c.lookupHandler(topic)
			}
		}()
	}
	wg.Wait()

	if _, ok := c.lookupHandler(topic); !ok {
		t.Fatal("注册后应能取到 handler")
	}
	if _, ok := c.lookupHandler("not-registered"); ok {
		t.Fatal("未注册的 topic 应返回 ok=false")
	}
}
