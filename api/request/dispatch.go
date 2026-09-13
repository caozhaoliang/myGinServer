package request

type NodeSaveReq struct {
	Id       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	Schedule string `json:"schedule"`
}

type LineSaveReq struct {
	Id       string `json:"id"`
	AheadId  string `json:"ahead_id"`
	BehindId string `json:"behind_id"`
	Type     string `json:"type"`
}
