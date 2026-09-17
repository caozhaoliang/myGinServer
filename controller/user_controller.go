package controller

import (
	"myGinServer/api/request"
	"myGinServer/api/response"
	"myGinServer/config"
	"myGinServer/internal/store"
	"myGinServer/models/article"
	"myGinServer/models/task"
	user2 "myGinServer/models/user"
	"myGinServer/service"
	"myGinServer/service/objectserver"
	"myGinServer/service/tenantserver"
	"myGinServer/service/userserver"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UserController struct {
	userServer    *userserver.UserServer
	tasksServer   *userserver.TasksServer
	articleServer *service.ArticleServer
	objectServer  *objectserver.ObjectServer
	tenantServer  *tenantserver.TenantServer
}

func NewUserController(db store.DBStore, cfg *config.Config) *UserController {
	tasksServer := userserver.NewTasksServer(db)
	userServer := userserver.NewUserServer(db)
	articleServer := service.NewArticleServer(db)
	objectServer := objectserver.NewObjectServer(&cfg.ObjectConfig)
	tenantServer := tenantserver.NewTenantServer(db)
	return &UserController{
		tasksServer:   tasksServer,
		userServer:    userServer,
		articleServer: articleServer,
		objectServer:  objectServer,
		tenantServer:  tenantServer,
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

// Profile 获取用户个人信息
// @Summary 获取用户个人信息
// @Description 根据用户ID获取用户的详细信息
// @Tags 用户相关
// @Accept json
// @Produce json
// @Success 200 {object} user2.User "用户信息"
// @Failure 400 {object} interface{} "请求参数错误"
// @Failure 500 {object} interface{} "服务器内部错误"
// @Router /user/profile [get]
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

// TenantList 获取当前用户参与的租户列表。
func (u *UserController) TenantList(c *gin.Context) {
	user, exists := c.Get("id")
	if !exists {
		SendError(c, http.StatusUnauthorized, errors.New("未登录"))
		return
	}
	tenants, err := u.tenantServer.List(c, user.(*user2.User).UserId)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, tenants)
}

// UpdateProfile 更新当前用户个人资料。
func (u *UserController) UpdateProfile(c *gin.Context) {
	var req request.UpdateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	user, exists := c.Get("id")
	if !exists {
		SendError(c, http.StatusUnauthorized, errors.New("未登录"))
		return
	}
	if err := u.userServer.UpdateProfile(c, user.(*user2.User).UserId, &req); err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
}

// ChangePassword 修改当前用户密码。
func (u *UserController) ChangePassword(c *gin.Context) {
	var req request.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	user, exists := c.Get("id")
	if !exists {
		SendError(c, http.StatusUnauthorized, errors.New("未登录"))
		return
	}
	if err := u.userServer.ChangePassword(c, user.(*user2.User).UserId, &req); err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
}

// UserList 分页查询用户列表（管理员）。
func (u *UserController) UserList(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	list, total, err := u.userServer.List(c, keyword, page, size)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, response.UserListResp{List: list, Total: total})
}

// CreateUser 创建用户（管理员）。
func (u *UserController) CreateUser(c *gin.Context) {
	var req request.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	id, err := u.userServer.Create(c, &req)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, id)
}

// UpdateUser 更新用户基础信息（管理员）。
func (u *UserController) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}
	var req request.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	if err := u.userServer.Update(c, id, &req); err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
}

// DisableUser 禁用用户（管理员，禁止禁用自己）。
func (u *UserController) DisableUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}
	user, exists := c.Get("id")
	if !exists {
		SendError(c, http.StatusUnauthorized, errors.New("未登录"))
		return
	}
	if err := u.userServer.Disable(c, id, user.(*user2.User).UserId); err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
}

// ResetUserPassword 重置用户密码（管理员）。
func (u *UserController) ResetUserPassword(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}
	var req request.ResetPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	if err := u.userServer.ResetPassword(c, id, req.NewPassword); err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, nil)
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
func (u *UserController) SaveArticle(c *gin.Context) {
	var req article.ArticleSaveReq
	err := c.ShouldBindJSON(&req)
	if err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	var article = article.ArticleVO{
		Id:        uuid.New().String(),
		ChannelId: req.ChannelId,
		Title:     req.Title,
		Status:    "draft",
		Cover:     req.Cover,
		Pubdate:   article.CustomTime(time.Now()),
	}
	err = u.articleServer.SaveArticle(c, &article)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, article.Id)
}

func (u *UserController) GetArticle(c *gin.Context) {
	var err error
	idParam := c.Param("id")
	if len(idParam) == 0 {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}
	detail, err := u.articleServer.ArticleDetail(c, idParam)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
		return
	}
	SendSuccess(c, detail)
}
func (u *UserController) UpdateArticle(c *gin.Context) {
	var articleVo = article.ArticleVO{}
	if err := c.ShouldBindJSON(&articleVo); err != nil {
		SendError(c, http.StatusBadRequest, err)
		return
	}
	if articleVo.Id == "" {
		SendError(c, http.StatusBadRequest, errors.New("ID为空"))
		return
	}
	err := u.articleServer.UpdateArticle(c, &articleVo)
	if err != nil {
		SendError(c, http.StatusInternalServerError, err)
	}
	SendSuccess(c, nil)
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
