package request

type NodeSaveReq struct {
	Id       string `json:"id"`
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Content  string `json:"content"`
	Type     string `json:"type" binding:"required,oneof=Virtual SQL Collect Sync"`
	Schedule string `json:"schedule" binding:"required"`
}

type LineSaveReq struct {
	Id       string `json:"id"`
	AheadId  string `json:"ahead_id" binding:"required"`
	BehindId string `json:"behind_id" binding:"required"`
	Type     string `json:"type" binding:"required,oneof=Dotted Solid"`
}

type DatasourceReq struct {
	Id      string `json:"id"`
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Type    string `json:"type" binding:"required"`
	ConnStr string `json:"conn_str" binding:"required"`
}
