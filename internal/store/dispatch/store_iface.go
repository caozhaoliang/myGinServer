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
	GetDatasource(ctx context.Context, project string, id string) (dispatch.Datasource, error)

	ExecQueueExists(ctx context.Context, project, runId string) (bool, error)
	SaveExecQueue(ctx context.Context, project string, entity dispatch.ExecQueue) error
	QueryExecQueue(ctx context.Context, project string, runId string) (*dispatch.ExecQueue, error)
	UpdateExecQueueResp(ctx context.Context, project, runId, resp, status string) error
}
