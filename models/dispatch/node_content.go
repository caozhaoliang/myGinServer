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

// ShellNodeContent 节点 content：包含一段在远程主机上执行的 shell 文本。
type ShellNodeContent struct {
	Shell string `json:"shell"`
}
