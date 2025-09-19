package controller

import (
	"myGinServer/internal/store"
	"myGinServer/models/article"
	"myGinServer/models/task"
	user2 "myGinServer/models/user"
	"myGinServer/service"
	"myGinServer/service/userserver"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserController struct {
	userServer    *userserver.UserServer
	tasksServer   *userserver.TasksServer
	articleServer *service.ArticleServer
}

func NewUserController(db store.DBStore) *UserController {
	tasksServer := userserver.NewTasksServer(db)
	userServer := userserver.NewUserServer(db)
	articleServer := service.NewArticleServer(db)
	return &UserController{
		tasksServer:   tasksServer,
		userServer:    userServer,
		articleServer: articleServer,
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

func (u *UserController) Profile(c *gin.Context) {
	user, exists := c.Get("id")
	if !exists {
		SendError(c, http.StatusInternalServerError, errors.New("用户不存在"))
		return
	}
	userInfo, err := u.userServer.Profile(c, user.(*user2.User).UserId)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, userInfo)
}

func (u *UserController) Tasks(c *gin.Context) {
	keyword, _ := c.GetQuery("keyword")
	tasks, err := u.tasksServer.TaskList(c, keyword)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, tasks)
}

func (u *UserController) DelTask(c *gin.Context) {
	var err error
	idParam := c.Param("id")
	if len(idParam) == 0 {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}

	err = u.tasksServer.TaskDel(c, idParam)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
}

func (u *UserController) SaveTask(c *gin.Context) {
	var taskReq task.Task
	if err := c.ShouldBindJSON(&taskReq); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	id, err := u.tasksServer.TaskSave(c, taskReq)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, id)
}

func (u *UserController) Channels(c *gin.Context) {
	channels, err := u.articleServer.Channels(c)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, channels)
}

func (u *UserController) Articles(c *gin.Context) {
	var req article.ArticlesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	articles, err := u.articleServer.Articles(c, &req)
	SendSuccess(c, articles)
}

func (u *UserController) DeleteArticle(c *gin.Context) {
	id, b := c.GetQuery("id")
	if !b {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}
	err := u.articleServer.DeleteArticle(c, id)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
}
