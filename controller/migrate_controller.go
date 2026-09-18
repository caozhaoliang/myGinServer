package controller

import (
	"errors"
	"net/http"

	"myGinServer/api/request"
	"myGinServer/service/dispatchserver"
	"myGinServer/service/migrate"

	"github.com/gin-gonic/gin"
)

// MigrateController 数据迁移接口：提供类 DataX 的 MySQL → MySQL 迁移能力。
// 源/目标连接与表名取自已保存的采集/同步节点 content（node_id），
// 接口只接收运行参数，不接收任何连接信息；仅需登录（不依赖租户中间件，
// 因为节点解析时后端自行从节点数据源与租户上下文取连接）。
type MigrateController struct {
	migrator   *migrate.Migrator
	nodeServer *dispatchserver.NodeServer
}

func NewMigrateController(nodeServer *dispatchserver.NodeServer) *MigrateController {
	return &MigrateController{
		migrator:   migrate.NewMigrator(),
		nodeServer: nodeServer,
	}
}

// Run 提交一个迁移任务（异步执行，幂等键 run_id）。
// 连接信息、源表/目标表名来自 node_id 指向的节点 content；
// 请求中的非空运行参数会覆盖节点解析出的默认值。
func (m *MigrateController) Run(c *gin.Context) {
	var req request.MigrateRunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	cfg, err := m.nodeServer.MigrateConfigFromNode(c.Request.Context(), req.NodeId)
	if err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	// 运行参数覆盖（连接与表名一律以节点 content 为准）
	if req.TargetTable != "" {
		cfg.TargetTable = req.TargetTable
	}
	if req.Where != "" {
		cfg.Where = req.Where
	}
	if req.BatchSize > 0 {
		cfg.BatchSize = req.BatchSize
	}
	if req.Mode != "" {
		cfg.Mode = req.Mode
	}
	if req.Channels > 0 {
		cfg.Channels = req.Channels
	}
	if req.CreateTableIfMissing != nil {
		cfg.CreateTableIfMissing = *req.CreateTableIfMissing
	}
	if err := m.migrator.Run(req.RunId, *cfg); err != nil {
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
