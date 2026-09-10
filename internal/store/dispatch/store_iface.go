package dispatch

import (
	"context"
	"myGinServer/models/dispatch"
)

type StoreIface interface {
	SaveNode(ctx context.Context, project string, node dispatch.Node) error
}
