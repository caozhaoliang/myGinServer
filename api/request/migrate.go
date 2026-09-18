package request

// DBEndpoint MySQL 连接信息（源 / 目标）
type DBEndpoint struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password"`
	Database string `json:"database" binding:"required"`
}

// MigrateRunReq 迁移任务提交请求：把源库一张表的数据迁移到目标库同名表。
// 幂等键 run_id：同 run_id 重复提交且任务未结束时直接返回已存在。
type MigrateRunReq struct {
	RunId  string `json:"run_id" binding:"required"`
	Source DBEndpoint `json:"source" binding:"required"`
	Target DBEndpoint `json:"target" binding:"required"`
	// Table 迁移源表名
	Table string `json:"table" binding:"required"`
	// TargetTable 目标表名（默认与 Table 相同；改名迁移时指定）
	TargetTable string `json:"target_table"`
	// Columns 目标列（顺序与源查询一致），空表示全列按源表列顺序
	Columns []string `json:"columns"`
	// Where 可选过滤条件（原样拼进 SELECT，如 "id > 1000"，勿带 WHERE 关键字）
	Where string `json:"where"`
	// BatchSize 每个事务写入的行数，默认 2000
	BatchSize int `json:"batch_size"`
	// Mode 写入模式：insert（默认，目标表需无冲突）/ replace（REPLACE INTO）/ upsert（ON DUPLICATE KEY UPDATE）
	Mode string `json:"mode" binding:"oneof=insert replace upsert"`
	// CreateTableIfMissing 目标表不存在时，用源表的 SHOW CREATE TABLE 自动建表
	CreateTableIfMissing bool `json:"create_table_if_missing"`
	// Channels 并发通道数（默认 1 串行）。>1 时按数字主键切分区间并行，
	// 源表无数字主键时自动回退串行并记录提示。
	Channels int `json:"channels"`
}

// MigrateStopReq 停止迁移任务
type MigrateStopReq struct {
	RunId string `json:"run_id" binding:"required"`
}
