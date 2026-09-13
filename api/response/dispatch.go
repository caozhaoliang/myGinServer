package response

import "myGinServer/api/request"

type Graph struct {
	Lines []request.LineSaveReq
	Nodes []request.NodeSaveReq
}
