// Package tenantctx 提供在 context.Context 中透传「租户库名」与「当前用户 ID」的辅助函数。
// 租户中间件在请求进入服务层之前把这些值写入请求 context，服务层统一从这里读取，
// 从而避免依赖 gin.Context：后台 cron 与单元测试传入的是纯 context.Context。
package tenantctx

import "context"

// ctxKey 使用私有类型作为 context key，避免与其它包的 key 冲突。
// 注意：不能用空结构体（type ctxKey struct{}）——Go 中所有空结构体字面量
// 都是同一个零值、相互相等，两个「不同的」ctxKey{} 实例会命中同一个 context key，
// 导致 TenantDB 与 UserID 互相覆盖（实测：WithUserID 会覆盖 WithTenantDB 写入的值）。
type tenantDBKeyType int
type userIDKeyType int

var (
	tenantDBKey tenantDBKeyType
	userIDKey   userIDKeyType
)

// WithTenantDB 在 ctx 中写入租户对应的数据库名，返回新的 context。
func WithTenantDB(ctx context.Context, db string) context.Context {
	return context.WithValue(ctx, tenantDBKey, db)
}

// TenantDB 从 ctx 中读取租户库名，缺失时返回空字符串（空字符串表示默认库）。
func TenantDB(ctx context.Context) string {
	if v, ok := ctx.Value(tenantDBKey).(string); ok {
		return v
	}
	return ""
}

// WithUserID 在 ctx 中写入当前登录用户 ID，返回新的 context。
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey, id)
}

// UserID 从 ctx 中读取当前登录用户 ID，缺失时返回空字符串。
func UserID(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}
