// Package tenant 定义租户与用户-租户关联相关的数据模型。
package tenant

import "database/sql"

// Tenant 租户表：按 database_name 分库，业务数据落在各自的库中。
type Tenant struct {
	TenantID     string       `db:"tenant_id" json:"tenant_id"`
	Name         string       `db:"name" json:"name"`
	Code         string       `db:"code" json:"code"`
	DatabaseName string       `db:"database_name" json:"database_name"`
	Status       string       `db:"status" json:"status"`
	CreatedOn    sql.NullTime `db:"created_on" json:"created_on"`
	CreatedBy    string       `db:"created_by" json:"created_by"`
}

// UserTenant 用户-租户关联表，记录用户在租户内的角色与默认租户标识。
type UserTenant struct {
	ID           int64        `db:"id" json:"id"`
	UserID       string       `db:"user_id" json:"user_id"`
	TenantID     string       `db:"tenant_id" json:"tenant_id"`
	RoleInTenant string       `db:"role_in_tenant" json:"role_in_tenant"`
	IsDefault    int          `db:"is_default" json:"is_default"`
	Status       string       `db:"status" json:"status"`
	CreatedOn    sql.NullTime `db:"created_on" json:"created_on"`
}

// TenantMember 用户可见的租户列表项：租户基本信息 + 该用户在租户内的角色。
type TenantMember struct {
	TenantID     string `db:"tenant_id" json:"tenant_id"`
	Name         string `db:"name" json:"name"`
	Code         string `db:"code" json:"code"`
	DatabaseName string `db:"database_name" json:"database_name"`
	Status       string `db:"status" json:"status"`
	RoleInTenant string `db:"role_in_tenant" json:"role_in_tenant"`
	IsDefault    int    `db:"is_default" json:"is_default"`
}
