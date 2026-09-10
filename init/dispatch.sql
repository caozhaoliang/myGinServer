package init

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
        KEY `idx_code` (`code`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点表';

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
