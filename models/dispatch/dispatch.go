package dispatch

import (
	"database/sql"
)

type NodeType string
type NodeStatus string
type LineType string

var (
	NodeVirtual NodeType = "Virtual"
	NodeSQL     NodeType = "SQL"
	NodeCollect NodeType = "Collect"
	NodeSync    NodeType = "Sync"

	NormalStatus NodeStatus = "Normal"
	DryRun       NodeStatus = "DryRun"
	StopRun      NodeStatus = "StopRun"
)

var (
	Dotted LineType = "Dotted" // 点虚线
	Solid  LineType = "Solid"  // 实线
)

type ExecStatus string

var (
	Pending ExecStatus = "pending"
	Loading ExecStatus = "loading"
	Running ExecStatus = "running"
	Success ExecStatus = "success"
	Failed  ExecStatus = "failed"
)

type Nodes struct {
	Id        string         `json:"id" gorm:"column:id;primaryKey;type:VARCHAR(64);not null;comment:主键ID"`                             // 主键ID
	Code      string         `json:"code" gorm:"column:code;type:VARCHAR(64);not null;default:;comment:节点编码"`                           // 节点编码
	Name      string         `json:"name" gorm:"column:name;type:VARCHAR(128);not null;default:;comment:节点名称"`                          // 节点名称
	Content   sql.NullString `json:"content" gorm:"column:content;type:MEDIUMTEXT;comment:节点内容"`                                        // 节点内容
	Schedule  string         `json:"schedule" gorm:"column:schedule;type:VARCHAR(64);not null;default:;comment:Cron表达式（调度时间）"`          // Cron表达式（调度时间）
	Type      NodeType       `json:"type" gorm:"column:type;type:VARCHAR(32);not null;default:;comment:节点类型：Virtual/SQL/Collect/Sync"`  // 节点类型：Virtual/SQL/Collect/Sync
	Status    NodeStatus     `json:"status" gorm:"column:status;type:VARCHAR(32);not null;default:;comment:节点状态：Normal/DryRun/StopRun"` // 节点状态：Normal/DryRun/StopRun
	Deleted   int32          `json:"deleted" gorm:"column:deleted;type:TINYINT(1);not null;default:0;comment:逻辑删除标识：0-未删除，1-已删除"`       // 逻辑删除标识：0-未删除，1-已删除
	CreatedOn sql.NullTime   `json:"created_on" gorm:"column:created_on;type:DATETIME;default:NULL;comment:创建时间"`                       // 创建时间
	CreatedBy sql.NullString `json:"created_by" gorm:"column:created_by;type:VARCHAR(64);default:NULL;comment:创建人ID"`                   // 创建人ID
}

func (n *Nodes) TableName() string {
	return "nodes"
}

type Line struct {
	Id        string         `json:"id" gorm:"column:id;primaryKey;type:VARCHAR(64);not null;comment:连线ID"`                       // 连线ID
	AheadId   string         `json:"ahead_id" gorm:"column:ahead_id;type:VARCHAR(64);not null;comment:前驱节点ID"`                    // 前驱节点ID
	BehindId  string         `json:"behind_id" gorm:"column:behind_id;type:VARCHAR(64);not null;comment:后继节点ID"`                  // 后继节点ID
	Type      LineType       `json:"type" gorm:"column:type;type:VARCHAR(20);not null;comment:连线类型: Dotted(点虚线), Solid(实线)"`      // 连线类型: Dotted(点虚线), Solid(实线)
	Deleted   int32          `json:"deleted" gorm:"column:deleted;type:TINYINT(1);not null;default:0;comment:逻辑删除标识：0-未删除，1-已删除"` // 逻辑删除标识：0-未删除，1-已删除
	CreatedOn sql.NullTime   `json:"created_on" gorm:"column:created_on;type:DATETIME;default:NULL;comment:创建时间"`                 // 创建时间
	CreatedBy sql.NullString `json:"created_by" gorm:"column:created_by;type:VARCHAR(64);default:NULL;comment:创建人ID"`             // 创建人ID
}

func (l *Line) TableName() string {
	return "line"
}

type Datasource struct {
	Id        string         `json:"id" gorm:"column:id;primaryKey;type:VARCHAR(64);not null;comment:ID"`               // ID
	Code      string         `json:"code" gorm:"column:code;type:VARCHAR(64);not null;default:;comment:编码"`             // 编码
	Name      string         `json:"name" gorm:"column:name;type:VARCHAR(128);not null;default:;comment:名称"`            // 名称
	Type      string         `json:"type" gorm:"column:type;type:VARCHAR(32);not null;default:mysql;comment:类型"`        //
	ConnStr   string         `json:"conn_str" gorm:"column:conn_str;type:VARCHAR(1024);not null;default:;comment:连接信息"` // 连接信息
	CreatedOn sql.NullTime   `json:"created_on" gorm:"column:created_on;type:DATETIME;default:NULL;comment:创建时间"`       // 创建时间
	CreatedBy sql.NullString `json:"created_by" gorm:"column:created_by;type:VARCHAR(64);default:NULL;comment:创建人ID"`   // 创建人ID
}

func (d *Datasource) TableName() string {
	return "datasource"
}

// 执行队列表
type ExecQueue struct {
	Id        string         `gorm:"column:id;type:varchar(64);primary_key;comment:ID" json:"id"`
	RunId     string         `gorm:"column:run_id;type:varchar(32);comment:任务运行ID;NOT NULL" json:"run_id"`
	Status    string         `gorm:"column:status;type:varchar(32);default:pending;comment:状态：pending、loading、running、success、failed;NOT NULL" json:"status"`
	Content   string         `gorm:"column:content;type:mediumtext;comment:运行内容" json:"content"`
	Response  string         `gorm:"column:response;type:mediumtext;comment:响应内容" json:"response"`
	CreatedOn sql.NullTime   `gorm:"column:created_on;type:datetime;comment:创建时间" json:"created_on"`
	CreatedBy sql.NullString `gorm:"column:created_by;type:varchar(64);comment:创建人ID" json:"created_by"`
}

func EntityFinished(status string) bool {
	return status == "success" || status == "failed"
}
func EntityAlreadyRun(status string) bool {
	return status == "running" || EntityFinished(status)
}

func (m *ExecQueue) TableName() string {
	return "exec_queue"
}
