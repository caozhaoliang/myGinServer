package dispatch

import (
	"context"
	"myGinServer/models/dispatch"
	"myGinServer/pkg/saas_db"
)

type Store struct {
	saas saas_db.SaaS
}

func NewDispatchStore(config *saas_db.DBConfig) StoreIface {

	return &Store{saas: saas_db.NewSaasDb(config)}
}

func (s *Store) SaveNode(ctx context.Context, project string, node dispatch.Nodes) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	r := db.Save(node)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func (s *Store) SaveLine(ctx context.Context, project string, line dispatch.Line) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	r := db.Save(line)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func (s *Store) NodeList(ctx context.Context, project string) ([]dispatch.Nodes, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return nil, err
	}
	var nodes []dispatch.Nodes
	r := db.Find(&nodes).Where(&dispatch.Nodes{Deleted: 0})
	if r.Error != nil {
		return nil, r.Error
	}
	return nodes, nil
}

func (s *Store) ListLines(ctx context.Context, project string) ([]dispatch.Line, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return nil, err
	}
	var lines []dispatch.Line
	r := db.Find(&lines).Where(&dispatch.Line{Deleted: 0})
	if r.Error != nil {
		return nil, r.Error
	}
	return lines, nil
}

func (s *Store) SaveDatasource(ctx context.Context, project string, req dispatch.Datasource) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	r := db.Save(&req)
	if r.Error != nil {
		return r.Error
	}
	return nil
}

func (s *Store) ListDatasource(ctx context.Context, project string) ([]dispatch.Datasource, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return nil, err
	}
	var ds []dispatch.Datasource
	err = db.Find(&ds).Error
	return ds, err
}
