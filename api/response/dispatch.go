package response

import "myGinServer/api/request"

type Graph struct {
	Lines []request.LineSaveReq `json:"lines"`
	Nodes []request.NodeSaveReq `json:"nodes"`
}

type MetaTables struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type MetaColumns struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Comment string `json:"comment"`
	Value   string `json:"value"` // 可选，分区字段 value 值

	PrimaryKeySeq int `json:"primary_key_seq"`
}
