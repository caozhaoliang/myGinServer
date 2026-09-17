-- 租户与用户-租户关联表，以及 users / exec_queue 的租户相关字段补充。
-- 默认租户与 admin 绑定由服务启动时的 EnsureDefaultTenant 幂等完成：
-- admin 的 user_id 是运行时生成的 UUID，无法在 SQL 中写死。

CREATE TABLE `tenant` (
    tenant_id VARCHAR(64) NOT NULL PRIMARY KEY COMMENT '租户ID（主键）',
    name VARCHAR(128) NOT NULL COMMENT '租户名称',
    code VARCHAR(64) NOT NULL UNIQUE COMMENT '租户编码（唯一）',
    database_name VARCHAR(64) NOT NULL UNIQUE COMMENT '租户对应的数据库名（仅小写字母/数字/下划线，后端生成）',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '租户状态（active：正常，disabled：禁用）',
    created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    created_by VARCHAR(64) NOT NULL DEFAULT '' COMMENT '创建人ID'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='租户表';

CREATE TABLE `user_tenant` (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '自增主键',
    user_id VARCHAR(64) NOT NULL COMMENT '用户ID',
    tenant_id VARCHAR(64) NOT NULL COMMENT '租户ID',
    role_in_tenant VARCHAR(20) NOT NULL DEFAULT 'member' COMMENT '租户内角色：owner/admin/member',
    is_default TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认租户：0-否，1-是',
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '关联状态（active：正常，disabled：禁用）',
    created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_user_tenant (user_id, tenant_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户-租户关联表';

ALTER TABLE users ADD COLUMN nickname VARCHAR(64) NOT NULL DEFAULT '' COMMENT '昵称' AFTER username;
ALTER TABLE users ADD COLUMN avatar VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像' AFTER nickname;

ALTER TABLE exec_queue ADD COLUMN tenant_db VARCHAR(64) NOT NULL DEFAULT '' COMMENT '所属租户库名' AFTER created_by;
INSERT IGNORE INTO users(user_id, username, nickname, avatar, password_hash, email, phone_num, role, status)
  VALUES('00000000-0000-0000-0000-000000000001',
       'admin',
       '管理员',
       '',
       '$2a$10$wQlLlI8W119xa04sOG25auHyC7iDbcYdtLiSFUdd7C.wG7weC3ri2',
       'admin@example.com',
       '',
       'admin',
       'active');

-- 默认租户（与 EnsureDefaultTenant 运行时种子一致）
  INSERT IGNORE INTO tenant(tenant_id, name, code, database_name, status, created_on, created_by)
  VALUES('default', '默认租户', 'default', 'mytest', 'active', NOW(), 'system');

  -- 将内置 admin 用户绑定为默认租户 owner（依赖上一步已插入的 admin 用户）
  INSERT IGNORE INTO user_tenant(user_id, tenant_id, role_in_tenant, is_default, status, created_on)
  VALUES('00000000-0000-0000-0000-000000000001', 'default', 'owner', 1, 'active', NOW());