package request

// MigrateRunReq 迁移任务提交请求（类 DataX，MySQL → MySQL）。
// 来源与目标的连接信息、表名一律取自已保存的采集/同步节点 content
// （node_id 指定），接口只传运行参数，不传任何连接信息。
type MigrateRunReq struct {
	RunId string `json:"run_id" binding:"required"`
	// NodeId 采集（Collect）/同步（Sync）节点 ID：后端按节点类型解析两端连接与表名
	NodeId string `json:"node_id" binding:"required"`
	// TargetTable 可选：覆盖节点 content 中的目标表名（如迁移到改名表）
	TargetTable string `json:"target_table"`
	// Where 可选过滤条件（原样拼进 SELECT，如 "id > 1000"，勿带 WHERE 关键字）
	Where string `json:"where"`
	// BatchSize 每个事务写入的行数，默认 2000
	BatchSize int `json:"batch_size"`
	// Mode 写入模式：insert（默认，目标表需无冲突）/ replace（REPLACE INTO）/ upsert（ON DUPLICATE KEY UPDATE）
	Mode string `json:"mode" binding:"omitempty,oneof=insert replace upsert"`
	// CreateTableIfMissing 目标表不存在时自动建表：节点 content 配置了 create_ddl 时默认已建，
	// 传 true 时即使无自定义 DDL 也会用源表 SHOW CREATE TABLE 建表；传 false 时关闭自动建表。
	CreateTableIfMissing *bool `json:"create_table_if_missing"`
	// Channels 并发通道数（默认 1 串行）。>1 时按数字主键切分区间并行，
	// 源表无数字主键时自动回退串行并记录提示。
	Channels int `json:"channels"`
}

// MigrateStopReq 停止迁移任务
type MigrateStopReq struct {
	RunId string `json:"run_id" binding:"required"`
}
