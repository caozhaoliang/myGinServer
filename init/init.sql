
CREATE TABLE users (
                       user_id VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '用户ID（主键）',
                       username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名（唯一）',
                       password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希（bcrypt/sha256等加密后存储）',
                       email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱（唯一，用于登录和找回密码）',
                       status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '账号状态（active：正常，disabled：禁用，locked：锁定）',
                       created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                       updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

create table `tasks` (
    id char(36) not null primary key,
    name varchar(1024) not null default '',
    des text,
    completed tinyint(2) not null default 0
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务表';

insert into tasks values (1, '吃饭', '吃饭的描述', 0),(2, '睡觉', '睡觉的描述', 0),(3, '打代码', '打代码的描述', 0);


CREATE TABLE IF NOT EXISTS channel (
    id VARCHAR(36) NOT NULL PRIMARY KEY,  -- 假设 ID 是 UUID 格式，长度 36
    name VARCHAR(100) NOT NULL,           -- 频道名称，非空
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '频道状态（active：正常，disabled：禁用）',
    des VARCHAR(255) default '',                             -- 描述，可空
    created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP         -- 创建时间，非空
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='频道表';

INSERT INTO channel (id, name, des,status, created_on)
VALUES (
           'c1f8d7e6-5a4b-3c2d-1e0f-9g8h7i6j5k4l',
           'GO',
           '用于讨论GO编程语言、开发工具等技术话题',
           'active',
           '2023-10-01 08:30:00'
       ),(
           'd2e3f4a5-b6c7-8d9e-0f1g-2h3i4j5k6l7m',
           'Python',
           '用于讨论Python编程语言、开发工具等技术话题',
          'active',
           '2023-10-02 14:15:00'
       );
