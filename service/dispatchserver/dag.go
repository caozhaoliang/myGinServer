package dispatchserver

import (
	"context"
	"database/sql"
	store "myGinServer/internal/store/dispatch"
	"myGinServer/models/dispatch"
	"myGinServer/utils"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type WarpNode struct {
	dispatch.Nodes
	instances []dispatch.NodeInstance
}
type DAGNode struct {
	Data WarpNode
	Ins  map[string]*DAGNode
	Outs map[string]*DAGNode
}

type NodeDAG struct {
	Nodes   map[string]*DAGNode
	BatchId string
}
type InstanceDAGBuilder struct {
	Id         string
	dag        NodeDAG
	store      store.StoreIface
	visitedMap map[string]struct{}
}

func NewInstanceDAGBuilder(id, batchId string, saas store.StoreIface) *InstanceDAGBuilder {
	return &InstanceDAGBuilder{
		Id:         id,
		store:      saas,
		dag:        NodeDAG{Nodes: make(map[string]*DAGNode), BatchId: batchId},
		visitedMap: map[string]struct{}{},
	}
}
func (n *DAGNode) AddIn(in *DAGNode) {
	n.Ins[in.Data.Id] = in
}

func (n *DAGNode) AddOut(out *DAGNode) {
	n.Outs[out.Data.Id] = out
}

type InstanceDependEntity map[string][]string

func (e InstanceDependEntity) Merge(r InstanceDependEntity) InstanceDependEntity {
	if len(r) == 0 {
		return e
	}
	for curInstanceID, upstreamInstanceIds := range e {
		if _, exists := r[curInstanceID]; exists {
			r[curInstanceID] = append(r[curInstanceID], upstreamInstanceIds...) // 直接append 是否需要去重？
			continue
		}
		r[curInstanceID] = upstreamInstanceIds
	}
	return r
}

// BuildInstanceDepend 生成实例依赖关系
func (n *DAGNode) BuildInstanceDepend() InstanceDependEntity {
	if len(n.Data.instances) == 0 || len(n.Ins) == 0 {
		return nil
	}
	result := make(map[string][]string)
	curNode := n.Data
	for _, upstreamNode := range n.Ins {
		upNode := upstreamNode.Data
		if len(upNode.instances) == 0 {
			continue
		}
		idx, pIdx := len(curNode.instances)-1, len(upNode.instances)-1
		for idx >= 0 && pIdx >= 0 {
			previous := upNode.instances[pIdx]
			curInstance := curNode.instances[idx]
			if curInstance.ExecuteTime.Time.Before(previous.ExecuteTime.Time) && pIdx > 0 {
				pIdx--
				continue
			}
			if _, ok := result[curNode.instances[idx].Id]; !ok {
				result[curNode.instances[idx].Id] = make([]string, 0)
			}
			result[curNode.instances[idx].Id] = append(result[curNode.instances[idx].Id], previous.Id)
			idx--
		}
	}
	return result
}

func (n *DAGNode) CreateInstance(lstExecutionTime []time.Time, batchId string) {
	if len(lstExecutionTime) == 0 {
		nodeInstanceID := uuid.New().String()
		node := dispatch.NodeInstance{
			Id:          nodeInstanceID,
			NodeId:      n.Data.Id,
			Name:        n.Data.Name,
			RunStyle:    string(dispatch.DryRun),
			Status:      string(dispatch.InstanceStatusNotReady),
			ExecuteTime: sql.NullTime{Time: time.Now(), Valid: true},
			StartTime:   sql.NullTime{}, // 写库的时候这个字段是NOT NULL
			EndTime:     sql.NullTime{},
			BatchId:     batchId,
			Index:       1,
		}

		n.Data.instances = append(n.Data.instances, node)
		return
	}
	for i, executionTime := range lstExecutionTime {
		nodeInstanceID := uuid.New().String()
		node := dispatch.NodeInstance{
			Id:          nodeInstanceID,
			NodeId:      n.Data.Id,
			Name:        n.Data.Name,
			RunStyle:    string(n.Data.Status),
			Status:      string(dispatch.InstanceStatusNotReady),
			ExecuteTime: sql.NullTime{Time: executionTime, Valid: true},
			StartTime:   sql.NullTime{}, // 写库的时候这个字段是NOT NULL
			EndTime:     sql.NullTime{},
			BatchId:     batchId,
			Index:       i + 1,
		}

		n.Data.instances = append(n.Data.instances, node)
	}
	return
}

func newNodes(node dispatch.Nodes) *DAGNode {
	return &DAGNode{
		Data: WarpNode{
			Nodes:     node,
			instances: make([]dispatch.NodeInstance, 0),
		},
		Ins:  make(map[string]*DAGNode),
		Outs: make(map[string]*DAGNode),
	}
}

func (i *InstanceDAGBuilder) buildDAGNode(node dispatch.Nodes) (nodes *DAGNode) {
	var (
		ok bool
	)
	if nodes, ok = i.dag.Nodes[node.Id]; !ok {
		nodes = newNodes(node)
		i.dag.Nodes[node.Id] = nodes
	}
	return
}

func (i *InstanceDAGBuilder) Build(lines []dispatch.Line) (NodeDAG, *dispatch.Nodes, error) {
	if len(i.dag.Nodes) == 0 {
		i.buildDAGNode(dispatch.Nodes{Id: i.Id})
	}
	for _, dep := range lines {
		behindNode := i.buildDAGNode(dispatch.Nodes{Id: dep.BehindId})
		aheadNode := i.buildDAGNode(dispatch.Nodes{Id: dep.AheadId})
		linkDAGNodes(aheadNode, behindNode)
	}
	var root = &dispatch.Nodes{}
	nodes, err := i.store.NodeList(context.TODO(), "")
	if err != nil {
		return NodeDAG{}, nil, err
	}
	mpNodes := make(map[string]dispatch.Nodes, len(nodes))
	for _, n := range nodes {
		mpNodes[n.Id] = n
	}
	for id, _ := range i.dag.Nodes {
		_, exists := mpNodes[id]
		if !exists {
			delete(i.dag.Nodes, id)
		}
	}
	for _, node := range nodes {
		if node.Id == i.Id {
			root = &node
		}
		i.replaceDAGNode(node)
	}
	for _, dep := range lines {
		behindNode, exist := i.dag.Nodes[dep.BehindId]
		aheadNode, exist2 := i.dag.Nodes[dep.AheadId]
		if !exist || !exist2 {
			continue
		}
		linkDAGNodes(aheadNode, behindNode)
	}
	return i.dag, root, nil
}

func (i *InstanceDAGBuilder) replaceDAGNode(node dispatch.Nodes) (nodes *DAGNode) {
	if _, ok := i.dag.Nodes[node.Id]; ok {
		nodes = newNodes(node)
		i.dag.Nodes[node.Id] = nodes
	}
	return
}

func linkDAGNodes(upstream *DAGNode, downstream *DAGNode) {
	upstream.AddOut(downstream)
	downstream.AddIn(upstream)
}

// --- 解析cron 表达式 ---

// bizDate 2006-01-02
func getStartTime(bizDate string) time.Time {
	var (
		err error
		t   time.Time
	)
	if t, err = time.ParseInLocation("2006-01-02", bizDate, time.Local); err == nil {
		return t.AddDate(0, 0, 1)
	}
	sYear, sMonth, sDay := time.Now().Local().Add(1 * time.Hour * 24).Date()
	return time.Date(sYear, sMonth, sDay, 0, 0, 0, 0, time.Local)
}
func (n *NodeDAG) Parser(bizDate string) error {
	start := getStartTime(bizDate)

	for _, node := range n.Nodes {
		// 区间为 [start, end)，右端直接取次日零点，由半开区间自然排除次日零点这一次触发；
		// 用 AddDate 而非 Add(24h)，跨夏令时切换时仍能稳定落在次日零点。
		times, err := utils.GetSchedulesBetween(node.Data.Schedule, start, start.AddDate(0, 0, 1))
		if err != nil {
			return errors.Wrap(err, "解析 cron表达式错误")
		}
		node.CreateInstance(times, n.BatchId)
	}
	return nil
}

// BuildInstanceDepend 为每个节点的实例生成上下游依赖关系
func (n *NodeDAG) BuildInstanceDepend() map[string][]string {
	result := make(map[string][]string)
	for _, node := range n.Nodes {
		result = node.BuildInstanceDepend().Merge(result)
	}
	return result
}
func (n *NodeDAG) GetInstanceList() []dispatch.NodeInstance {
	nodeInstances := make([]dispatch.NodeInstance, 0)
	for _, node := range n.Nodes {
		nodeInstances = append(nodeInstances, node.Data.instances...)
	}
	return nodeInstances
}

// GetInstanceById 取节点的首个实例。第二个返回值为 false 表示节点不存在，
// 或其 cron 表达式在本批次窗口内合法地没有产生任何实例（如 0 0 * * 1 落在非周一），
// 调用方必须显式处理，不能拿着零值继续往下走。
func (n *NodeDAG) GetInstanceById(id string) (dispatch.NodeInstance, bool) {
	node, ok := n.Nodes[id]
	if !ok || len(node.Data.instances) == 0 {
		return dispatch.NodeInstance{}, false
	}
	return node.Data.instances[0], true
}
