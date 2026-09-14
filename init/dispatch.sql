
-- 创建节点表
CREATE TABLE `nodes` (
        `id` VARCHAR(64) NOT NULL COMMENT '主键ID',
        `code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '节点编码',
        `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '节点名称',
        `content` MEDIUMTEXT COMMENT '节点内容',
        `schedule` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'Cron表达式（调度时间）',
        `type` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '节点类型：Virtual/SQL/Collect/Sync',
        `status` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '节点状态：Normal/DryRun/StopRun',
        `deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '逻辑删除标识：0-未删除，1-已删除',
        `created_on` DATETIME DEFAULT NULL COMMENT '创建时间',
        `created_by` VARCHAR(64) DEFAULT NULL COMMENT '创建人ID',
        PRIMARY KEY (`id`),
        KEY `idx_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点表';
replace into nodes(id, code, name, content, schedule, type, status, deleted, created_on, created_by)
values("615966d0-af61-11f1-8f44-866b84541a87","ROOT_NODE_CODE","根节点","",
       "0 0 * * *", "Virtual", "DryRun", 0, now(), "admin")

-- 创建连线表
CREATE TABLE `line` (
    `id` VARCHAR(64) NOT NULL COMMENT '连线ID',
    `ahead_id` VARCHAR(64) NOT NULL COMMENT '前驱节点ID',
    `behind_id` VARCHAR(64) NOT NULL COMMENT '后继节点ID',
    `type` VARCHAR(20) NOT NULL COMMENT '连线类型: Dotted(点虚线), Solid(实线)',
    `deleted` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '逻辑删除标识：0-未删除，1-已删除',
    `created_on` DATETIME DEFAULT NULL COMMENT '创建时间',
    `created_by` VARCHAR(64) DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_ahead_behind_id` (`ahead_id`,`behind_id`),
    KEY `idx_behind_id` (`behind_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='连线信息表';

CREATE TABLE IF NOT EXISTS `datasource` (
    `id` VARCHAR(64) NOT NULL COMMENT 'ID',
    `code` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '编码',
    `name` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '名称',
    `type` VARCHAR(64)  NOT NULL DEFAULT 'mysql' COMMENT '数据源类型',
    `conn_str` varchar(1024) NOT NULL DEFAULT '' COMMENT '连接信息',
    `created_on` DATETIME DEFAULT NULL COMMENT '创建时间',
    `created_by` VARCHAR(64) DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    KEY `idx_code` (`code`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据源表';
insert into datasource(id,code,name,`type`,conn_str,created_on,created_by)
values ("615966d0-af61-11f1-8f44-866b84548888", "default","default", "mysql","{}",now(),"admin")

CREATE TABLE IF NOT EXISTS `exec_queue` (
    `id` VARCHAR(64) NOT NULL COMMENT 'ID',
    `run_id` VARCHAR(32) NOT NULL COMMENT '任务运行ID',
    `status` VARCHAR(32) NOT NULL DEFAULT 'pending' COMMENT '状态：pending\loading\running\success\failed'
    `content` MEDIUMTEXT COMMENT '运行内容',
    `response` MEDIUMTEXT COMMENT '响应内容',
    `created_on` DATETIME DEFAULT NULL COMMENT '创建时间',
    `created_by` VARCHAR(64) DEFAULT NULL COMMENT '创建人ID',
    PRIMARY KEY (`id`),
    UNIQUE KEY unq_idx(`run_id`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='执行队列表';
