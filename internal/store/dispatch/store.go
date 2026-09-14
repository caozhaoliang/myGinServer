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
	err = db.Save(&node).Error
	return err
}

func (s *Store) SaveLine(ctx context.Context, project string, line dispatch.Line) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	return db.Save(&line).Error
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

func (s *Store) GetDatasource(ctx context.Context, project string, id string) (dispatch.Datasource, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return dispatch.Datasource{}, err
	}
	var ds dispatch.Datasource
	if err := db.Where("id = ?", id).First(&ds).Error; err != nil {
		return dispatch.Datasource{}, err
	}
	return ds, nil
}

func (s *Store) ExecQueueExists(ctx context.Context, project, runId string) (bool, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Model(&dispatch.ExecQueue{}).Where("run_id = ?", runId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) SaveExecQueue(ctx context.Context, project string, entity dispatch.ExecQueue) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	err = db.Save(&entity).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) QueryExecQueue(ctx context.Context, project string, runId string) (*dispatch.ExecQueue, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return nil, err
	}
	var queue dispatch.ExecQueue
	err = db.Where("run_id=?", runId).First(&queue).Error
	if err != nil {
		return nil, err
	}
	return &queue, nil
}

func (s *Store) UpdateExecQueueResp(ctx context.Context, project, runId, resp, status string) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	err = db.Model(&dispatch.ExecQueue{}).Where("run_id=?", runId).
		Updates(&dispatch.ExecQueue{Status: status, Response: resp}).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) InstanceExists(ctx context.Context, project, batchId string) (bool, error) {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return false, err
	}
	var count int64
	err = db.Model(&dispatch.NodeInstance{}).Where("batch_id = ?", batchId).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) BatchCreateInstance(ctx context.Context, project string,
	dep []dispatch.InstanceLine, instance []dispatch.NodeInstance) error {
	db, err := s.saas.GetDB(ctx, project)
	if err != nil {
		return err
	}
	tx := db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if len(dep) > 0 {
		err = tx.Model(&dispatch.InstanceLine{}).CreateInBatches(dep, 1000).Error
		if err != nil {
			return err
		}
	}
	if len(instance) > 0 {
		err = tx.Model(&dispatch.NodeInstance{}).CreateInBatches(instance, 1000).Error
	}
	return err
}
