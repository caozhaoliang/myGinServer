package controller

import (
	user2 "myGinServer/internal/user"
	"myGinServer/service/userserver"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userServer *userserver.UserServer
}

func NewUserController(userServer *userserver.UserServer) *UserController {
	return &UserController{
		userServer: userServer,
	}
}

func (u *UserController) Register(c *gin.Context) {
	var userReq struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// 绑定并验证请求数据
	if err := c.ShouldBindJSON(&userReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "无效的请求数据: " + err.Error(),
		})
		return
	}
	var user = &user2.User{
		Username:     userReq.Username,
		Email:        userReq.Email,
		PasswordHash: userReq.Password,
	}
	userId, err := u.userServer.Register(user)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}

	SendSuccess(c, userId)
}
