package request

type NodeSaveReq struct {
	// Id 为必填：保存逻辑是「按主键 upsert」，Id 为空时会以空字符串作为主键 INSERT，
	// 第二次提交即主键冲突。新增时请由前端生成 UUID。
	Id       string `json:"id" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Content  string `json:"content"`
	Type     string `json:"type" binding:"required,oneof=Virtual SQL Collect Sync"`
	Schedule string `json:"schedule" binding:"required"`
}

type LineSaveReq struct {
	// 同 NodeSaveReq.Id：必填，新增时由前端生成 UUID。
	Id       string `json:"id" binding:"required"`
	AheadId  string `json:"ahead_id" binding:"required"`
	BehindId string `json:"behind_id" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=Dotted Solid"`
}

type DatasourceReq struct {
	// 同 NodeSaveReq.Id：必填，新增时由前端生成 UUID。
	Id      string `json:"id" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
	ConnStr string `json:"conn_str" binding:"required"`
}

// ----测试运行 ----

// TestRunSqlReq 测试运行请求体
type TestRunSqlReq struct {
	RunId  string                 `json:"run_id" binding:"required"`
	Sql    string                 `json:"sql" binding:"required"`
	Params map[string]interface{} `json:"params" binding:"required"`
	Type   string                 `json:"type"`
}

type InstanceCreateReq struct {
	Id      string `json:"id"`       // 节点 id
	Project string `json:"project"`  // 项目名称
	BizDate string `json:"biz_date"` // 业务日期
	TestRun bool   `json:"test_run"` // 测试运行
}

func (i InstanceCreateReq) BatchId() string {
	return i.Project + i.BizDate
}
