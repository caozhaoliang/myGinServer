package tool

import (
	"myGinServer/internal/store"
	user2 "myGinServer/internal/user"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
)

var jwtKey = []byte("928df7424055d5be39c5a455969de75f563f53a33a842cf6bc995a2fffa22062")

type JwtAuthMiddleware struct {
	Middleware *jwt.GinJWTMiddleware
}

const (
	jwtIdentityKey = "id"
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
		Timeout:     time.Hour,
		MaxRefresh:  time.Hour,
		IdentityKey: jwtIdentityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*user2.User); ok {
				return jwt.MapClaims{
					jwtIdentityKey: v.UserId,
					jwtNameKey:     v.Username,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &user2.User{
				UserId:   claims[jwtIdentityKey].(string),
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
		//Authorizator: func(data interface{}, c *gin.Context) bool {
		//	if v, ok := data.(*CustomClaims); ok && v.UserID == "admin" {
		//		return true
		//	}
		//	return false
		//},
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
		Middleware: authMiddleware,
	}, err
}
