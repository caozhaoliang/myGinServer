
CREATE TABLE users (
                       user_id VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '用户ID（主键）',
                       username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名（唯一）',
                       password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希（bcrypt/sha256等加密后存储）',
                       email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱（唯一，用于登录和找回密码）',
                       status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '账号状态（active：正常，disabled：禁用，locked：锁定）',
                       created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                       updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';