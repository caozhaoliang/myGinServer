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
	g *gin.Engine
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
func NewRouter(controller *controller.UserController, db store.DBStore) *Router {
	route := &Router{
		g: gin.New(),
	}
	r := gin.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	pprof.Register(r)
	logger := tool.InitLogger()
	r.Use(RecoveryWithLogger(logger), gin.Logger())
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

	authed := r.Group("/auth")
	authed.POST("/login", jwtMiddleware.Middleware.LoginHandler)
	return route
}

func (r *Router) Start(address string) error {
	return r.g.Run(address)
}
