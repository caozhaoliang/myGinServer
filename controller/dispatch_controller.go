package controller

import (
	"context"
	"myGinServer/api/request"
	"myGinServer/config"
	"myGinServer/internal/store/delayqueue"
	"myGinServer/internal/store/dispatch"
	"myGinServer/internal/workflow"
	mdispatch "myGinServer/models/dispatch"
	"myGinServer/pkg/saas_db"
	"myGinServer/service/dispatchserver"
	"net/http"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

type DispatchController struct {
	nodeServer *dispatchserver.NodeServer
}

func NewDispatchController(ctx context.Context, config *config.Config) *DispatchController {
	// saaS := saas_db.NewSaaS()
	conf := &saas_db.DBConfig{
		Database:     config.DBConfig.DbName,
		Host:         config.DBConfig.DbHost,
		Password:     config.DBConfig.DbPassword,
		Port:         config.DBConfig.DbPort,
		User:         config.DBConfig.DbUser,
		MaxIdleConns: 2,
		MaxOpenConns: 2,
	}
	iStore := dispatch.NewDispatchStore(conf)
	db, err := saas_db.CreateEngine(conf.Database, conf)
	if err != nil {
		panic(err)
	}
	producer := delayqueue.NewProducer(db)
	consumer := delayqueue.NewConsumer(db)
	runtime := workflow.NewNodeRuntime(iStore, producer, consumer)
	// 必须先 Dispatch（内部 Register 注册 handler）再 StartWorkers：
	// worker 的 ticker 如果先跑起来，启动瞬间队列里已有的到期消息会因为取不到 handler 被直接判 failed。
	err = runtime.Dispatch(mdispatch.NodeInstanceTopic)
	if err != nil {
		panic(err)
	}
	consumer.StartWorkers(ctx, 4)
	nodeServer := dispatchserver.NewNodeServer(iStore, producer)
	nodeServer.Dispatch(ctx)
	return &DispatchController{nodeServer: nodeServer}
}

func (s *DispatchController) SaveNode(c *gin.Context) {
	var req request.NodeSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求数据: " + err.Error(),
		})
		return
	}
	err := s.nodeServer.SaveNode(c.Request.Context(), &req)
	if err != nil {
		SendError(c, 500, err)
		return
	}
	SendSuccess(c, "ok")
}

func (s *DispatchController) SavaLine(c *gin.Context) {
	var req request.LineSaveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, errors.New("无效的请求参数"))
		return
	}
	err := s.nodeServer.SaveLine(c.Request.Context(), &req)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
	} else {
		SendSuccess(c, "ok")
	}
	return
}
func (s *DispatchController) DeleteLine(c *gin.Context) {
	type DeleteLineReq struct {
		Id string `json:"id" form:"id"`
	}
	var req DeleteLineReq
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	err := s.nodeServer.DeleteLine(c.Request.Context(), req.Id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, "ok")
}

func (s *DispatchController) DeleteNode(c *gin.Context) {
	type DeleteLineReq struct {
		Id string `json:"id" form:"id"`
	}
	var req DeleteLineReq
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	err := s.nodeServer.DeleteNode(c.Request.Context(), req.Id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, "ok")
}

func (s *DispatchController) Graph(c *gin.Context) {

	r, err := s.nodeServer.Graph(c.Request.Context())
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, &r)
}
func (s *DispatchController) TestRun(c *gin.Context) {
	var req request.TestRunSqlReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	err := s.nodeServer.TestRun(c.Request.Context(), req)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, "测试运行失败"))
		return
	}
	SendSuccess(c, "ok")
}

func (s *DispatchController) QueryResult(c *gin.Context) {
	type QueryResultReq struct {
		RunId string `json:"run_id" form:"run_id"`
	}
	var req QueryResultReq
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	r, err := s.nodeServer.QueryTestResult(c.Request.Context(), req.RunId)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, "运行结果获取失败"))
		return
	}
	SendSuccess(c, r)
}

// ------datasource ------

func (s *DispatchController) SaveDatasource(c *gin.Context) {
	var req request.DatasourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, errors.New("获取请求参数失败"))
		return
	}
	err := s.nodeServer.SaveDatasource(c.Request.Context(), req)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, "保存数据源失败"))
		return
	}
	SendSuccess(c, "ok")
}

func (s *DispatchController) ListDatasource(c *gin.Context) {
	datasource, err := s.nodeServer.ListDatasource(c.Request.Context())
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, "获取数据源列表失败"))
		return
	}
	SendSuccess(c, datasource)
}
func (s *DispatchController) MetaTables(c *gin.Context) {

	type MetaTableReq struct {
		DatasourceId string `json:"ds_id" form:"ds_id"`
	}
	var req MetaTableReq
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, errors.Wrap(err, "获取数据源id失败"))
		return
	}
	tables, err := s.nodeServer.MetaTables(c.Request.Context(), req.DatasourceId)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, ""))
		return
	}
	SendSuccess(c, tables)
}
func (s *DispatchController) MetaColumns(c *gin.Context) {
	type MetaColumnReq struct {
		DatasourceId string `json:"ds_id" form:"ds_id"`
		Table        string `json:"table" form:"table"`
	}
	var req MetaColumnReq
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, errors.Wrap(err, "获取数据源id失败"))
		return
	}
	tables, err := s.nodeServer.MetaColumns(c.Request.Context(), req.DatasourceId, req.Table)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, ""))
		return
	}
	SendSuccess(c, tables)
}

func (s *DispatchController) OdsTables(c *gin.Context) {
	tables, err := s.nodeServer.OdsTables(c.Request.Context())
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, ""))
		return
	}
	SendSuccess(c, tables)
}
func (s *DispatchController) OdsColumns(c *gin.Context) {
	type OdsMetaColumnReq struct {
		Table string `json:"table" form:"table"`
	}
	var req OdsMetaColumnReq
	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, errors.Wrap(err, "获取参数失败"))
		return
	}
	tables, err := s.nodeServer.OdsColumns(c.Request.Context(), req.Table)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, ""))
		return
	}
	SendSuccess(c, tables)
}
