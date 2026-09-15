package workflow

import "myGinServer/models/dispatch"

type RuntimeCtx struct {
	node     dispatch.Nodes
	instance dispatch.NodeInstance
}
