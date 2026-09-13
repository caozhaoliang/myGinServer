package dispatchserver

import (
	"context"
	"database/sql"
	"myGinServer/api/request"
	mdispatch "myGinServer/models/dispatch"
	"time"
)

func (n *NodeServer) SaveDatasource(ctx context.Context, req request.DatasourceReq) error {
	err := n.store.SaveDatasource(ctx, "", mdispatch.Datasource{
		Id:        req.Id,
		Code:      req.Code,
		Name:      req.Name,
		Type:      req.Type,
		ConnStr:   req.ConnStr,
		CreatedOn: sql.NullTime{Time: time.Now()},
		CreatedBy: sql.NullString{String: "admin"},
	})

	return err
}

func (n *NodeServer) ListDatasource(ctx context.Context) ([]request.DatasourceReq, error) {
	datasource, err := n.store.ListDatasource(ctx, "")
	if err != nil {
		return nil, err
	}
	var res []request.DatasourceReq
	for _, ds := range datasource {
		res = append(res, request.DatasourceReq{
			Id:      ds.Id,
			Code:    ds.Code,
			Name:    ds.Name,
			Type:    ds.Type,
			ConnStr: ds.ConnStr,
		})
	}
	return res, nil
}
