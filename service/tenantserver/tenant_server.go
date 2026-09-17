// Package tenantserver 提供租户相关业务逻辑。
package tenantserver

import (
	"context"

	"myGinServer/internal/store"
	"myGinServer/models/tenant"
)

// TenantServer 租户服务。
type TenantServer struct {
	dbStore store.DBStore
}

// NewTenantServer 创建租户服务实例。
func NewTenantServer(dbStore store.DBStore) *TenantServer {
	return &TenantServer{dbStore: dbStore}
}

// List 返回当前用户参与的所有租户及其角色。
func (t *TenantServer) List(ctx context.Context, userId string) ([]tenant.TenantMember, error) {
	return t.dbStore.ListUserTenants(ctx, userId)
}
