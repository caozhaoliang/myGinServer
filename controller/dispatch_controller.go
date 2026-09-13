package controller

import (
	"errors"
	"myGinServer/api/request"
	"myGinServer/config"
	"myGinServer/internal/store/dispatch"
	"myGinServer/pkg/saas_db"
	"myGinServer/service/dispatchserver"
	"net/http"

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
