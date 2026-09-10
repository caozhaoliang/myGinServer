package dispatch

import "database/sql"

type NodeType string
type NodeStatus string
type LineType string

var (
	NodeVirtual NodeType
	NodeSQL     NodeType
	NodeCollect NodeType
	NodeSync    NodeType

	NormalStatus NodeStatus
	DryRun       NodeStatus
	StopRun      NodeStatus
)

var (
	Dotted LineType // 点虚线
	Solid  LineType // 实线
)

type Node struct {
	Id        string         `json:"id" db:"id"`
	Code      string         `json:"code" db:"code"`
	Name      string         `json:"name" db:"name"`
	Content   string         `json:"content" db:"content"`
	Schedule  string         `json:"schedule" db:"schedule"` // cron 表达式
	Type      NodeType       `json:"type" db:"type"`
	Status    NodeStatus     `json:"status" db:"status"`
	Deleted   bool           `json:"deleted" db:"deleted"`
	CreatedOn sql.NullTime   `json:"created_on" db:"created_on"`
	CreatedBy sql.NullString `json:"created_by" db:"created_by"`
}

type Line struct {
	Id        string         `json:"id" db:"id"`
	AheadId   string         `json:"ahead_id" db:"ahead_id"` //对应节点ID
	BehindId  string         `json:"behind_id" db:"behind_id"`
	Type      LineType       `json:"type" db:"type"`
	Deleted   bool           `json:"deleted" db:"deleted"`
	CreatedOn sql.NullTime   `json:"created_on" db:"created_on"`
	CreatedBy sql.NullString `json:"created_by" db:"created_by"`
}
