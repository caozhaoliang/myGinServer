package dispatch

type SqlNodeContent struct {
	Sql   string         `json:"sql" db:"sql"`
	Param map[string]any `json:"param" db:"param"`
}

type Column struct {
	Name      string `json:"name"`
	DataType  string `json:"data_type"`
	IsPrimary int8   `json:"is_primary"`
}

type CollectSource struct {
	DatasourceId string   `json:"datasource_id" db:"datasource_id"`
	Table        string   `json:"table" db:"table"`
	Database     string   `json:"database" db:"database"`
	Columns      []Column `json:"columns" db:"columns"`
	CustomSQL    string   `json:"custom_sql" db:"custom_sql"`
}
type CollectTarget struct {
	Table     string   `json:"table"`
	CreateDDL string   `json:"create_ddl"`
	Columns   []Column `json:"columns" db:"columns"`
	IsCover   int8     `json:"is_cover" db:"is_cover"`
}

type CollectNodeContent struct {
	Source CollectSource `json:"source"`
	Target CollectTarget `json:"target"`
}

// SyncNodeContent 同步节点 content：源与目标均为数据源引用，
// 是类 DataX 迁移（migrate）配置的载体，两端连接信息从数据源 ID 解析。
type SyncNodeContent struct {
	Source CollectSource `json:"source"`
	Target SyncTarget    `json:"target"`
}

// SyncTarget 同步目标：比采集目标多了数据源 ID 与可选库名覆盖。
type SyncTarget struct {
	DatasourceId string   `json:"datasource_id"` // 目标数据源 ID（必填）
	Database     string   `json:"database"`      // 可选：覆盖数据源自带库名
	Table        string   `json:"table"`
	CreateDDL    string   `json:"create_ddl"`
	Columns      []Column `json:"columns"`
	IsCover      int8     `json:"is_cover"` // 是否覆盖写入：1=replace，0=insert（接口 mode 显式指定时以接口为准）
}

// ShellNodeContent 节点 content：包含一段在远程主机上执行的 shell 文本。
type ShellNodeContent struct {
	Shell string `json:"shell"`
}
