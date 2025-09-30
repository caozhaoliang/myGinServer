
CREATE TABLE `users` (
                       user_id VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '用户ID（主键）',
                       username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名（唯一）',
                       password_hash VARCHAR(255) NOT NULL COMMENT '密码哈希（bcrypt/sha256等加密后存储）',
                       email VARCHAR(100) NOT NULL UNIQUE COMMENT '邮箱（唯一，用于登录和找回密码）',
                       status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '账号状态（active：正常，disabled：禁用，locked：锁定）',
                       created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                       updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';
alter table users add column phone_num varchar(64) not null default '' comment '手机号' after email;
alter table users add column role varchar(36) not null default 'register' comment '角色：admin、guest、register、vip、vip2' after email;

create table `tasks` (
    id char(36) not null primary key,
    name varchar(1024) not null default '',
    des text,
    completed tinyint(2) not null default 0
 ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='任务表';

insert into tasks values (1, '吃饭', '吃饭的描述', 0),(2, '睡觉', '睡觉的描述', 0),(3, '打代码', '打代码的描述', 0);


CREATE TABLE `channel` (
    id VARCHAR(36) NOT NULL PRIMARY KEY,  -- 假设 ID 是 UUID 格式，长度 36
    name VARCHAR(100) NOT NULL,           -- 频道名称，非空
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT '频道状态（active：正常，disabled：禁用）',
    des VARCHAR(255) default '',                             -- 描述，可空
    created_on DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP         -- 创建时间，非空
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='频道表';

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

CREATE TABLE `articles` (
    id VARCHAR(36) NOT NULL PRIMARY KEY COMMENT '文章唯一标识',
    title VARCHAR(255) NOT NULL COMMENT '文章标题',
    cover VARCHAR(1024) COMMENT '文章封面图片URL',
    channel_id VARCHAR(36) NOT NULL COMMENT '频道ID',
    status VARCHAR(20) NOT NULL COMMENT '文章状态（如：draft-草稿、published-已发布、deleted-已删除）',
    pubdate DATETIME COMMENT '发布时间（未发布时可为空）',
    view_count INT NOT NULL DEFAULT 0 COMMENT '浏览次数',
    comment_count INT NOT NULL DEFAULT 0 COMMENT '评论次数',
    like_count INT NOT NULL DEFAULT 0 COMMENT '点赞次数',
    INDEX idx_status (status),
    INDEX idx_pubdate (pubdate)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';

ALTER TABLE articles
    ADD COLUMN content TEXT COMMENT '文章内容' AFTER title;

insert into articles values (
    'a1b2c3d4-e5f6-7g8h-9i0j-1k2l3m4n5o6p',
    'Go 语言简介',
    '',
    'https://example.com/go-cover.png',
    'c1f8d7e6-5a4b-3c2d-1e0f-9g8h7i6j5k4l',
    'published',
    '2023-10-01 08:30:00',
    100,
    10,
    5
),(
    uuid(),
    'Python 语言简介',
    '',
    'http://localhost:9000/imagebucket/5q1z5tmizh.png',
    'd2e3f4a5-b6c7-8d9e-0f1g-2h3i4j5k6l7m',
    'published',
    '2025-10-01 08:30:00',
    199,
    103,
    54
);

CREATE TABLE `message` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `form_id` varchar(64) NOT NULL COMMENT '发送者ID（与用户表关联）',
    `target_id` varchar(64) NOT NULL COMMENT '接收者ID（与用户表关联）',
    `type` varchar(20) NOT NULL COMMENT '消息类型（如群聊 私聊 广播等）',
    `media` varchar(20) NOT NULL COMMENT '消息类型（如text/image/file等）',
    `content` text DEFAULT NULL COMMENT '消息内容（扩展字段，实际场景通常需要）',
    `pic` varchar(64) DEFAULT NULL COMMENT '图片',
    `url` varchar(64) DEFAULT NULL COMMENT '连接',
    `desc` varchar(512) DEFAULT NULL COMMENT '描述',
    `send_time` bigint(20) NOT NULL DEFAULT 0 COMMENT '发送时间戳（秒级）',
    `is_read` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
    PRIMARY KEY (`id`),
    KEY `idx_form_target` (`form_id`, `target_id`),
    KEY `idx_target_time` (`target_id`, `send_time`),
    KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';


CREATE TABLE `message` (
`id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
`form_id` varchar(64) NOT NULL COMMENT '发送者ID（与用户表关联）',
`target_id` varchar(64) NOT NULL COMMENT '接收者ID（与用户表关联）',
`type` varchar(20) NOT NULL COMMENT '消息类型（如群聊 私聊 广播等）',
`media` varchar(20) NOT NULL COMMENT '消息类型（如text/image/file等）',
`content` text DEFAULT NULL COMMENT '消息内容（扩展字段，实际场景通常需要）',
`pic` varchar(64) DEFAULT NULL COMMENT '图片',
`url` varchar(64) DEFAULT NULL COMMENT '连接',
`desc` varchar(512) DEFAULT NULL COMMENT '描述',
`send_time` bigint(20) NOT NULL DEFAULT 0 COMMENT '发送时间戳（秒级）',
`is_read` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否已读：0-未读，1-已读',
PRIMARY KEY (`id`),
KEY `idx_form_target` (`form_id`, `target_id`),
KEY `idx_target_time` (`target_id`, `send_time`),
KEY `idx_type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='消息表';

create table `contact` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `owner_id` varchar(32) not null default '' comment '关系的拥有者',
    `target_id` varchar(32) not null default '' comment '关系的对象',
    `type` varchar(16) not null default '' comment '关系类型',
    `desc` varchar(64) not null default '' comment '关系描述',
    primary key (`id`),
    key `idx_owner`(`owner_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='关系表';


create table `group_basic` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增主键',
    `name` varchar(64) not null default '' comment '分组名称',
    `owner_id` varchar(32) not null default '' comment '分组的拥有者',
    `icon` varchar(32) not null default '' comment '图标',
    `type` varchar(16) not null default '' comment '分组类型',
    `desc` varchar(64) not null default '' comment '分组描述',
    primary key (`id`),
    key `idx_owner`(`owner_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='分组表';