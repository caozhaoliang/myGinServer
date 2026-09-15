-- 节点坐标字段迁移脚本（面向已存在的库）
-- 新库直接使用 dispatch.sql 中的 CREATE TABLE（已包含 pos_x/pos_y 两列），无需执行本文件。
-- 老库请手动执行一次，为 nodes 表补充画布坐标列；列允许为 NULL，旧数据保持 NULL，前端据此退回网格布局。

ALTER TABLE `nodes`
    ADD COLUMN `pos_x` DOUBLE NULL DEFAULT NULL COMMENT '节点横坐标（画布位置）' AFTER `schedule`,
    ADD COLUMN `pos_y` DOUBLE NULL DEFAULT NULL COMMENT '节点纵坐标（画布位置）' AFTER `pos_x`;
