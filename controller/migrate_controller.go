package controller

import (
	"errors"
	"net/http"

	"myGinServer/api/request"
	"myGinServer/service/migrate"

	"github.com/gin-gonic/gin"
)

// MigrateController 数据迁移接口：提供类 DataX 的 MySQL → MySQL 迁移能力。
// 源/目标连接由调用方在请求中提供，接口仅需登录（不依赖租户库）。
type MigrateController struct {
	migrator *migrate.Migrator
}

func NewMigrateController() *MigrateController {
	return &MigrateController{migrator: migrate.NewMigrator()}
}

// Run 提交一个迁移任务（异步执行，幂等键 run_id）
func (m *MigrateController) Run(c *gin.Context) {
	var req request.MigrateRunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	cfg := migrate.Config{
		Source: migrate.Endpoint{
			Host: req.Source.Host, Port: req.Source.Port,
			User: req.Source.User, Password: req.Source.Password,
			Database: req.Source.Database,
		},
		Target: migrate.Endpoint{
			Host: req.Target.Host, Port: req.Target.Port,
			User: req.Target.User, Password: req.Target.Password,
			Database: req.Target.Database,
		},
		Table:                req.Table,
		TargetTable:          req.TargetTable,
		Columns:              req.Columns,
		Where:                req.Where,
		BatchSize:            req.BatchSize,
		Mode:                 req.Mode,
		CreateTableIfMissing: req.CreateTableIfMissing,
		Channels:             req.Channels,
	}
	if err := m.migrator.Run(req.RunId, cfg); err != nil {
		if errors.Is(err, migrate.ErrTaskExists) {
			SendError(c, http.StatusBadRequest, err)
			return
		}
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, "ok")
}

// Status 查询迁移任务状态
func (m *MigrateController) Status(c *gin.Context) {
	runId := c.Query("run_id")
	if runId == "" {
		SendError(c, http.StatusBadRequest, errors.New("缺少参数 run_id"))
		return
	}
	st, ok := m.migrator.Status(runId)
	if !ok {
		SendError(c, http.StatusNotFound, errors.New("任务不存在"))
		return
	}
	SendSuccess(c, st)
}

// Stop 取消一个运行中的迁移任务
func (m *MigrateController) Stop(c *gin.Context) {
	runId := c.Query("run_id")
	if runId == "" {
		SendError(c, http.StatusBadRequest, errors.New("缺少参数 run_id"))
		return
	}
	if err := m.migrator.Stop(runId); err != nil {
		if errors.Is(err, migrate.ErrTaskNotFound) {
			SendError(c, http.StatusNotFound, err)
			return
		}
		SendError(c, http.StatusBadRequest, err)
		return
	}
	SendSuccess(c, "ok")
}
