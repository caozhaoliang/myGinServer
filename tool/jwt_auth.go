package tool

import (
	"myGinServer/controller"
	"myGinServer/internal/store"
	user2 "myGinServer/models/user"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var jwtKey = []byte("928df7424055d5be39c5a455969de75f563f53a33a842cf6bc995a2fffa22062")

type JwtAuthMiddleware struct {
	Middleware *jwt.GinJWTMiddleware
	db         store.DBStore
}

const (
	JwtIdentityKey = "id"
	jwtNameKey     = "username"
	RootUser       = "admin"
	RootPassword   = "bigdata@2019!"
)

type user struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password"`
}

func NewJwtAuthMiddleware(db store.DBStore) (*JwtAuthMiddleware, error) {

	// the jwt Middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "",
		Key:         jwtKey,
		Timeout:     time.Hour * 8,
		MaxRefresh:  time.Hour,
		IdentityKey: JwtIdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(user2.User); ok {
				return jwt.MapClaims{
					JwtIdentityKey: v.UserId,
					jwtNameKey:     v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &user2.User{
				UserId:   claims[JwtIdentityKey].(string),
				Username: claims[jwtNameKey].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginReq user
			if err := c.ShouldBindJSON(&loginReq); err != nil {
				return nil, err
			}
			userInfo, err := db.CheckUserInfo(c, loginReq.Username, loginReq.PasswordHash)
			if err != nil {
				return nil, jwt.ErrFailedAuthentication
			}
			return userInfo, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// todo 这里可以根据data中的信息做一些更加细粒度的权限资源校验
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},

		TokenLookup: "header:Authorization, cookie:tk",
		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",
		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	return &JwtAuthMiddleware{
		db:         db,
		Middleware: authMiddleware,
	}, err
}

func (j *JwtAuthMiddleware) CallbackHandler(c *gin.Context) {
	fnHandler := func(userName, email string) {
		var err error
		userId := uuid.New().String()
		u := user2.User{UserId: userId, Email: email, Username: userName}
		defer func() {
			if err != nil {
				controller.SendError(c, 500, err)
			}
		}()
		userInfo, err := j.db.GetUserByName(c, userName)
		if err != nil {
			return
		}
		if len(userInfo.UserId) == 0 {
			// 用户不存在，保存用户
			err = j.db.SaveUser(c, u)
			if err != nil {
				return
			}
		} else {
			u.UserId = userInfo.UserId
		}
		// 保存用户完成，开始生成token
		token, _, err := j.Middleware.TokenGenerator(&u)
		if err != nil {
			return
		}
		controller.SendSuccess(c, token)
	}
	CallbackCheck(c, fnHandler)
}
