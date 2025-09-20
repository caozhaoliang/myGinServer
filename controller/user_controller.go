package controller

import (
	"myGinServer/config"
	"myGinServer/internal/store"
	"myGinServer/models/article"
	"myGinServer/models/task"
	user2 "myGinServer/models/user"
	"myGinServer/service"
	"myGinServer/service/objectserver"
	"myGinServer/service/userserver"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserController struct {
	userServer    *userserver.UserServer
	tasksServer   *userserver.TasksServer
	articleServer *service.ArticleServer
	objectServer  *objectserver.ObjectServer
}

func NewUserController(db store.DBStore, cfg *config.Config) *UserController {
	tasksServer := userserver.NewTasksServer(db)
	userServer := userserver.NewUserServer(db)
	articleServer := service.NewArticleServer(db)
	objectServer := objectserver.NewObjectServer(&cfg.ObjectConfig)
	return &UserController{
		tasksServer:   tasksServer,
		userServer:    userServer,
		articleServer: articleServer,
		objectServer:  objectServer,
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

func (u *UserController) PresignedUpload(c *gin.Context) {
	filename := c.Query("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}
	url, err := u.objectServer.Cli.PreSignedPutObject(filename, time.Minute*60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	SendSuccess(c, url.String())
}
