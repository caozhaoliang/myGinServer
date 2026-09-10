package dispatchserver

import (
	"context"
	"myGinServer/api/request"
	"myGinServer/internal/store/dispatch"
)

type NodeServer struct {
	store dispatch.StoreIface
}

func NewNodeServer(store dispatch.StoreIface) *NodeServer {
	return &NodeServer{store: store}
}

func (n *NodeServer) SaveNode(ctx context.Context, req *request.NodeSaveReq) error {
	// todo 根据入参进行节点校验，并更新或写入到数据库。

	return nil
}

func (n *NodeServer) SaveLine(ctx context.Context, req *request.LineSaveReq) error {
	// todo 新增或者更新，执行前需要判断是否成环。 utils.graph工具包

	return nil
}
