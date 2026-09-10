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

	return &Store{saas: saas_db.NewSaaS(config)}
}

func (s *Store) SaveNode(ctx context.Context, project string, node dispatch.Node) error {
	_, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}

	return nil
}
