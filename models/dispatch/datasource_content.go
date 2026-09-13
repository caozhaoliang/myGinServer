package dispatch

type MysqlDatasourceContent struct {
	User     string `json:"user"`
	Passwd   string `json:"passwd"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Schema   string `json:"schema"`
}
