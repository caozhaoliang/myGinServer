package router

import (
	"myGinServer/controller"
	"myGinServer/internal/store"
	"myGinServer/tool"
	"net/http"
	"net/http/httputil"
	"runtime/debug"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Router struct {
	r *gin.Engine
}

func RecoveryWithLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rval := recover(); rval != nil {
				debug.PrintStack()
				httprequest, _ := httputil.DumpRequest(c.Request, false)
				logger.WithFields(logrus.Fields{
					"uri":     c.Request.RequestURI,
					"request": string(httprequest),
				}).Error(rval)

				c.String(http.StatusInternalServerError, "InternalServerError")
				c.Abort()
			}
		}()

		c.Next()
	}
}
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许所有来源（生产环境建议指定具体域名）
		c.Header("Access-Control-Allow-Origin", "*")
		// 允许的请求头
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		// 允许的请求方法
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH, DELETE")
		// 允许前端获取的头信息
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		// 是否允许后续请求携带认证信息（cookies）
		c.Header("Access-Control-Allow-Credentials", "true")

		// 处理预检请求（OPTIONS方法）
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
func NewRouter(controller *controller.UserController, db store.DBStore) *Router {
	route := &Router{}
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	pprof.Register(r)
	logger := tool.InitLogger()
	r.Use(RecoveryWithLogger(logger), gin.Logger()).Use(Cors())
	r.POST("/register", controller.Register)

	jwtMiddleware, err := tool.NewJwtAuthMiddleware(db)
	if err != nil {
		panic(err)
	}
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hello world",
		})
	})
	r.GET("/callback", jwtMiddleware.CallbackHandler)
	r.POST("/auth/login", jwtMiddleware.Middleware.LoginHandler)
	api := r.Group("/api", jwtMiddleware.Middleware.MiddlewareFunc())
	{
		api.GET("/user/profile", controller.Profile)
		api.GET("/task/list", controller.Tasks)
		api.DELETE("/task/del/:id", controller.DelTask)
		api.POST("/task/add", controller.SaveTask)

		api.GET("/channels", controller.Channels)
		api.GET("/articles", controller.Articles)
		api.DELETE("/article", controller.DeleteArticle)
		api.POST("/article/save", controller.SaveArticle)
		api.POST("/article/update", controller.UpdateArticle)
		api.GET("/article/:id", controller.GetArticle)
	}
	apiObject := r.Group("/api/object", jwtMiddleware.Middleware.MiddlewareFunc())
	{
		apiObject.PUT("/presigned-upload-url", controller.PresignedUpload)
	}
	route.r = r
	return route
}

func (r *Router) Start(address string) error {
	return r.r.Run(address)
}
