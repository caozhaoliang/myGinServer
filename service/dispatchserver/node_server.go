package dispatchserver

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"myGinServer/api/request"
	"myGinServer/api/response"
	"myGinServer/internal/store/delayqueue"
	"myGinServer/internal/store/dispatch"
	mdispatch "myGinServer/models/dispatch"
	"myGinServer/utils"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	ErrLineCycle = errors.Errorf("成环异常")
)

type NodeServer struct {
	store    dispatch.StoreIface
	producer *delayqueue.Producer
}

func NewNodeServer(store dispatch.StoreIface, producer *delayqueue.Producer) *NodeServer {

	return &NodeServer{store: store, producer: producer}
}

func (n *NodeServer) SaveNode(ctx context.Context, req *request.NodeSaveReq) error {
	// 根据入参进行节点校验，并更新或写入到数据库。
	if err := utils.ValidateCronExpr(req.Schedule); err != nil {
		return errors.Wrapf(err, "未知的cron表达式:%s", req.Schedule)
	}
	err := n.store.SaveNode(ctx, "", mdispatch.Nodes{
		Id:        req.Id,
		Code:      req.Code,
		Name:      req.Name,
		Content:   sql.NullString{String: req.Content, Valid: true},
		Schedule:  req.Schedule,
		Type:      mdispatch.NodeType(req.Type),
		Status:    mdispatch.NormalStatus,
		Deleted:   0,
		CreatedOn: sql.NullTime{Time: time.Now(), Valid: true},
		CreatedBy: sql.NullString{String: "admin", Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (n *NodeServer) DeleteNode(ctx context.Context, nodeId string) error {
	return n.store.DeleteNode(ctx, "", nodeId)
}

func (n *NodeServer) DeleteLine(ctx context.Context, lineId string) error {
	return n.store.DeleteLine(ctx, "", lineId)
}

func (n *NodeServer) SaveLine(ctx context.Context, req *request.LineSaveReq) error {
	// 新增或者更新，执行前需要判断是否成环。 utils.graph工具包
	lines, err := n.store.ListLines(ctx, "")
	if err != nil {
		return err
	}
	graph := NewGraph(lines)
	isCycle := graph.wouldCreateCycleLocked(req.AheadId, req.BehindId)
	if isCycle {
		return errors.Wrapf(ErrLineCycle, "%s->%s", req.AheadId, req.BehindId)
	}
	err = n.store.SaveLine(ctx, "", mdispatch.Line{
		Id:        req.Id,
		AheadId:   req.AheadId,
		BehindId:  req.BehindId,
		Type:      mdispatch.LineType(req.Type),
		Deleted:   0,
		CreatedOn: sql.NullTime{Time: time.Now(), Valid: true},
		CreatedBy: sql.NullString{String: "admin", Valid: true},
	})
	return err
}

func (n *NodeServer) Graph(ctx context.Context) (response.Graph, error) {
	nodes, err := n.store.NodeList(ctx, "")
	if err != nil {
		return response.Graph{}, err
	}
	lists, err := n.store.ListLines(ctx, "")
	if err != nil {
		return response.Graph{}, err
	}
	nodeReq := []request.NodeSaveReq{}
	lineReq := []request.LineSaveReq{}
	for _, node := range nodes {
		nodeReq = append(nodeReq, request.NodeSaveReq{
			Id:       node.Id,
			Code:     node.Code,
			Name:     node.Name,
			Content:  node.Content.String,
			Type:     string(node.Type),
			Schedule: node.Schedule,
		})
	}
	for _, line := range lists {
		lineReq = append(lineReq, request.LineSaveReq{
			Id:       line.Id,
			AheadId:  line.AheadId,
			BehindId: line.BehindId,
			Type:     string(line.Type),
		})
	}

	return response.Graph{Nodes: nodeReq, Lines: lineReq}, nil
}

func (n *NodeServer) TestRun(ctx context.Context, req request.TestRunSqlReq) error {
	// 1.根据runId 做幂等校验 2.base64解码Sql字段 3.SQL参数替换 4.写入待执行队列(exec_queue表)
	exists, err := n.store.ExecQueueExists(ctx, "", req.RunId)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(req.Sql)
	if err != nil {
		return errors.Wrapf(err, "SQL 解码失败")
	}
	template, err := utils.RenderTemplate(string(decoded), req.Params)
	if err != nil {
		return errors.Wrapf(err, "参数替换失败")
	}
	err = n.store.SaveExecQueue(ctx, "", mdispatch.ExecQueue{
		Id:        uuid.New().String(),
		RunId:     req.RunId,
		Status:    string(mdispatch.Pending),
		Content:   fmt.Sprintf("{\"sql\":\"%s\"}", template),
		Response:  "{}",
		CreatedOn: sql.NullTime{time.Now(), true},
		CreatedBy: sql.NullString{"admin", true},
	})
	if err != nil {
		return err
	}
	err = n.SendEntity(ctx, req.RunId, template)
	if err != nil {
		// 补偿：上面已向 exec_queue 落了一条 pending 记录，但消息没能进入内存队列，
		// 它永远不会被执行，前端轮询 QueryResult 会一直拿到空对象、无法终止。
		// 这里把它置为 failed 并写入失败原因，让轮询能正常结束。
		//
		// 必须用「脱离取消」的 context：投递失败的常见原因正是 ctx 被取消或超时，
		// 若沿用原 ctx，这条补偿写入自身也会被一并取消掉。
		markCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
		defer cancel()
		// Response 必须是合法 JSON：QueryTestResult 对 finished 状态会直接 json.Unmarshal。
		payload, _ := json.Marshal(response.TestRunResp{Msg: "投递执行队列失败: " + err.Error(), Sql: template})
		if markErr := n.store.UpdateExecQueueResp(markCtx, "", req.RunId, string(payload), string(mdispatch.Failed)); markErr != nil {
			// 补偿也失败，两个错误一并抛出，避免其中之一被静默吞掉
			return fmt.Errorf("投递执行队列失败: %v；标记 exec_queue 为 failed 亦失败: %v", err, markErr)
		}
		return err
	}
	return nil
}

func (n *NodeServer) QueryTestResult(ctx context.Context, runId string) (response.TestRunResp, error) {
	// 获取exec_queue中的response字段并解析 返回结果。
	queue, err := n.store.QueryExecQueue(ctx, "", runId)
	if err != nil {
		return response.TestRunResp{}, err
	}
	var resp response.TestRunResp
	if mdispatch.EntityFinished(queue.Status) {
		err = json.Unmarshal([]byte(queue.Response), &resp)
		if err != nil {
			return response.TestRunResp{}, err
		}
	}
	return resp, nil
}
