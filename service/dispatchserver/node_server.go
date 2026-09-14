package dispatchserver

import (
	"context"
	"database/sql"
	"myGinServer/api/request"
	"myGinServer/api/response"
	"myGinServer/internal/store/dispatch"
	mdispatch "myGinServer/models/dispatch"
	"myGinServer/utils"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrLineCycle = errors.Errorf("成环异常")
)

type NodeServer struct {
	store dispatch.StoreIface
}

func NewNodeServer(store dispatch.StoreIface) *NodeServer {
	return &NodeServer{store: store}
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
