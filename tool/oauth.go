package tool

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

// 配置OAuth2.0参数
var (
	// 请替换为你的GitHub OAuth应用信息
	clientID     = "your_client_id"
	clientSecret = "your_client_secret"
	redirectURL  = "http://localhost:8081/callback"

	// 配置OAuth2.0端点
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"}, // 请求的权限范围
		Endpoint:     github.Endpoint,        // GitHub的OAuth端点
	}

	// 用于防止CSRF攻击的随机字符串
	stateString = "random-state-string-123"
)

// GitHub用户信息结构体
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

func CallbackCheck(c *gin.Context, fn func(user, email string)) {
	state := c.Query("state")
	if state != stateString {
		c.String(http.StatusBadRequest, "state不匹配，可能存在CSRF攻击")
		return
	}

	// 获取授权码
	code := c.Query("code")
	if code == "" {
		c.String(http.StatusBadRequest, "未获取到授权码")
		return
	}

	// 使用授权码获取访问令牌
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "获取令牌失败: %v", err)
		return
	}

	// 使用令牌获取用户信息
	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.String(http.StatusInternalServerError, "获取用户信息失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 解析用户信息
	var gitUser GitHubUser
	if err = json.NewDecoder(resp.Body).Decode(&gitUser); err != nil {
		c.String(http.StatusInternalServerError, "解析用户信息失败: %v", err)
		return
	}
	// 保存用户信息到数据库 2 生成当前应用的token并返回
	fn(gitUser.Login, gitUser.Email)

}
