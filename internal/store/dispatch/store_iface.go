package dispatch

import (
	"context"
	"myGinServer/models/dispatch"
)

type StoreIface interface {
	SaveNode(ctx context.Context, project string, node dispatch.Nodes) error
	SaveLine(ctx context.Context, project string, line dispatch.Line) error
	NodeList(ctx context.Context, project string) ([]dispatch.Nodes, error)
	ListLines(ctx context.Context, project string) ([]dispatch.Line, error)

	SaveDatasource(ctx context.Context, project string, req dispatch.Datasource) error
	ListDatasource(ctx context.Context, project string) ([]dispatch.Datasource, error)
}
