package tool

import (
	"net/http"

	"myGinServer/internal/store"
	user2 "myGinServer/models/user"
	"myGinServer/pkg/tenantctx"

	"github.com/gin-gonic/gin"
)

// TenantMiddleware 租户中间件：校验请求头 X-Tenant-Id，并在公共库中确认当前用户是该租户成员，
// 然后把「租户库名」与「用户 ID」写入请求 context，供服务层进行分库路由。
func TenantMiddleware(db store.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := c.GetHeader("X-Tenant-Id")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "缺少租户标识 X-Tenant-Id"})
			c.Abort()
			return
		}
		identity, ok := c.Get(JwtIdentityKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "未登录"})
			c.Abort()
			return
		}
		u := identity.(*user2.User)
		tenantInfo, err := db.GetTenantMember(c, u.UserId, tenantID)
		if err != nil || tenantInfo.TenantID == "" {
			c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "非该租户成员"})
			c.Abort()
			return
		}
		ctx := tenantctx.WithTenantDB(c.Request.Context(), tenantInfo.DatabaseName)
		ctx = tenantctx.WithUserID(ctx, u.UserId)
		c.Request = c.Request.WithContext(ctx)
		c.Set("tenant_id", tenantID)
		c.Set("tenant_db", tenantInfo.DatabaseName)
		c.Next()
	}
}

// RequireAdmin 管理员权限中间件：校验当前登录用户角色为 admin。
func RequireAdmin(db store.DBStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, ok := c.Get(JwtIdentityKey)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "未登录"})
			c.Abort()
			return
		}
		u := identity.(*user2.User)
		userInfo, err := db.GetUser(c, u.UserId)
		if err != nil || userInfo.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"code": http.StatusForbidden, "message": "无权限"})
			c.Abort()
			return
		}
		c.Next()
	}
}
