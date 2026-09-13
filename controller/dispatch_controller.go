package controller

import (
	"myGinServer/api/request"
	"myGinServer/config"
	"myGinServer/internal/store/dispatch"
	"myGinServer/pkg/saas_db"
	"myGinServer/service/dispatchserver"
	"net/http"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
)

type DispatchController struct {
	nodeServer *dispatchserver.NodeServer
}

func NewDispatchController(config *config.Config) *DispatchController {
	// saaS := saas_db.NewSaaS()
	iStore := dispatch.NewDispatchStore(&saas_db.DBConfig{
		Database:     config.DBConfig.DbName,
		Host:         config.DBConfig.DbHost,
		Password:     config.DBConfig.DbPassword,
		Port:         config.DBConfig.DbPort,
		User:         config.DBConfig.DbUser,
		MaxIdleConns: 2,
		MaxOpenConns: 2,
	})
	nodeServer := dispatchserver.NewNodeServer(iStore)

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
	err := s.nodeServer.SaveNode(c, &req)
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
	err := s.nodeServer.SaveLine(c, &req)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
	} else {
		SendSuccess(c, "ok")
	}
	return
}

func (s *DispatchController) Graph(c *gin.Context) {

	r, err := s.nodeServer.Graph(c)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, &r)
}

func (s *DispatchController) SaveDatasource(c *gin.Context) {
	var req request.DatasourceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, errors.New("获取请求参数失败"))
		return
	}
	err := s.nodeServer.SaveDatasource(c, req)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, "保存数据源失败"))
		return
	}
	SendSuccess(c, "ok")
}

func (s *DispatchController) ListDatasource(c *gin.Context) {
	datasource, err := s.nodeServer.ListDatasource(c)
	if err != nil {
		SendError(c, http.StatusInternalServerError, errors.Wrap(err, "获取数据源列表失败"))
		return
	}
	SendSuccess(c, datasource)
}
