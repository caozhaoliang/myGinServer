package dispatchserver

import (
	"testing"
	"time"

	"myGinServer/models/dispatch"
)

// TestParser_RootNodeMidnight 回归测试：根节点调度在午夜整点时，Parser 必须能生成实例。
//
// 此前的缺陷：GetSchedulesBetween 的取值循环直接用窗口起点 start 起步，
// 而 cron 的 Next 返回的是「严格大于」入参的时间点，导致恰好落在 start 上的触发
// （0 0 * * * 正好落在窗口左侧的午夜整点）被整条吞掉，根节点实例数为 0，
// 随后 GetInstanceById 取 instances[0] 时越界 panic。
func TestParser_RootNodeMidnight(t *testing.T) {
	const rootID = "root-node"
	// 业务日期 2026-09-15 → 调度窗口为次日 2026-09-16 整日
	const bizDate = "2026-09-15"

	dag := NodeDAG{
		Nodes: map[string]*DAGNode{
			rootID: newNodes(dispatch.Nodes{
				Id:       rootID,
				Code:     "ROOT_NODE_CODE",
				Name:     "根节点",
				Schedule: "0 0 * * *",
				Type:     dispatch.NodeVirtual,
				Status:   dispatch.DryRun,
			}),
		},
		BatchId: "test-batch",
	}

	if err := dag.Parser(bizDate); err != nil {
		t.Fatalf("Parser 返回错误: %v", err)
	}

	root := dag.Nodes[rootID]
	if len(root.Data.instances) != 1 {
		t.Fatalf("根节点生成 %d 个实例，期望恰好 1 个", len(root.Data.instances))
	}

	got := root.Data.instances[0]
	want := time.Date(2026, 9, 16, 0, 0, 0, 0, time.Local)
	if !got.ExecuteTime.Valid || !got.ExecuteTime.Time.Equal(want) {
		t.Fatalf("实例预期执行时间 = %v，期望 %v", got.ExecuteTime.Time, want)
	}
	if got.Status != string(dispatch.InstanceStatusNotReady) {
		t.Fatalf("实例初始状态 = %s，期望 %s", got.Status, dispatch.InstanceStatusNotReady)
	}
	if got.BatchId != dag.BatchId {
		t.Fatalf("实例批次 ID = %s，期望 %s", got.BatchId, dag.BatchId)
	}

	// 实例存在后，GetInstanceById 应返回 ok=true，而不是越界 panic
	if _, ok := dag.GetInstanceById(rootID); !ok {
		t.Fatal("GetInstanceById 未取到根节点实例")
	}
}

// TestParser_EmptyScheduleWindow 校验「窗口内合法为空」时不再 panic，而是返回 ok=false。
// TestParser_EmptyScheduleWindow 校验「窗口内合法为空」时不再 panic，而是生成一条空跑（DryRun）实例
// （c9d5944 引入的设计：无合法触发时也保留实例，供调度链路走空跑状态）。
// 0 0 * * 1 表示每周一零点；2026-09-15 作为业务日期解析出的窗口（09-16 周三）内没有周一。
func TestParser_EmptyScheduleWindow(t *testing.T) {
	const nodeID = "monday-node"

	dag := NodeDAG{
		Nodes: map[string]*DAGNode{
			nodeID: newNodes(dispatch.Nodes{
				Id:       nodeID,
				Name:     "周一节点",
				Schedule: "0 0 * * 1",
				Type:     dispatch.NodeVirtual,
				Status:   dispatch.DryRun,
			}),
		},
		BatchId: "test-batch",
	}

	if err := dag.Parser("2026-09-15"); err != nil {
		t.Fatalf("Parser 返回错误: %v", err)
	}

	instances := dag.Nodes[nodeID].Data.instances
	if n := len(instances); n != 1 {
		t.Fatalf("窗口内无周一，期望生成 1 条空跑实例，实际 %d 个", n)
	}
	if got := instances[0].RunStyle; got != string(dispatch.DryRun) {
		t.Fatalf("空跑实例 RunStyle 应为 DryRun，实际 %q", got)
	}
	if got := instances[0].Status; got != string(dispatch.InstanceStatusNotReady) {
		t.Fatalf("空跑实例 Status 应为 NotReady，实际 %q", got)
	}
	if _, ok := dag.GetInstanceById(nodeID); !ok {
		t.Fatal("存在空跑实例时 GetInstanceById 应返回 ok=true")
	}
}
