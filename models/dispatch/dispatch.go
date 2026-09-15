package dispatch

import (
	"database/sql"
)

const (
	NodeInstanceTopic = "instance_execute"
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
	PosX      *float64       `json:"x" gorm:"column:pos_x;type:DOUBLE;default:NULL;comment:节点横坐标（画布位置）"`                             // 节点横坐标（画布位置）
	PosY      *float64       `json:"y" gorm:"column:pos_y;type:DOUBLE;default:NULL;comment:节点纵坐标（画布位置）"`                             // 节点纵坐标（画布位置）
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

// -------- 实例 -----
type InstanceStatus string

var (
	InstanceStatusNotReady InstanceStatus = "NotReady"
	InstanceStatusWaiting  InstanceStatus = "Waiting"
	InstanceStatusRunning  InstanceStatus = "Running"
	InstanceStatusSuccess  InstanceStatus = "Success"
	InstanceStatusFailure  InstanceStatus = "Failure"
	InstanceStatusAbort    InstanceStatus = "Abort"
	InstanceStatusTimeOut  InstanceStatus = "TimeOut"
)

// 节点实例表
type NodeInstance struct {
	Id          string       `gorm:"column:id;type:varchar(64);primary_key;comment:ID" json:"id"`
	NodeId      string       `gorm:"column:node_id;type:varchar(32);comment:节点id;NOT NULL" json:"node_id"`
	Name        string       `gorm:"column:name;type:varchar(32);comment:名称;NOT NULL" json:"name"`
	Index       int          `gorm:"column:index;type:tinyint(4);default:0;comment:批次内节点实例索引;NOT NULL" json:"index"`
	ExecuteTime sql.NullTime `gorm:"column:execute_time;type:timestamp;comment:预期执行时间;NOT NULL" json:"execute_time"`
	// start_time / end_time 可空：实例在创建时尚未开始执行，结束时间更是要等执行完成才回填。
	StartTime sql.NullTime   `gorm:"column:start_time;type:timestamp;comment:开始执行时间" json:"start_time"`
	EndTime   sql.NullTime   `gorm:"column:end_time;type:timestamp;comment:执行结束时间" json:"end_time"`
	Status    string         `gorm:"column:status;type:varchar(32);default:NotReady;comment:状态;NOT NULL" json:"status"`
	BatchId   string         `gorm:"column:batch_id;type:varchar(32);comment:批次ID;NOT NULL" json:"batch_id"`
	CreatedOn sql.NullTime   `gorm:"column:created_on;type:datetime;comment:创建时间" json:"created_on"`
	CreatedBy sql.NullString `gorm:"column:created_by;type:varchar(64);comment:创建人ID" json:"created_by"`
}

func (m *NodeInstance) TableName() string {
	return "node_instance"
}

// InstanceLine 实例连线表
type InstanceLine struct {
	Id        string         `gorm:"column:id;type:varchar(64);primary_key;comment:ID" json:"id"`
	Ahead     string         `gorm:"column:ahead;type:varchar(32);comment:ahead;NOT NULL" json:"ahead"`
	Behind    string         `gorm:"column:behind;type:varchar(32);comment:behind;NOT NULL" json:"behind"`
	BatchId   string         `gorm:"column:batch_id;type:varchar(32);comment:batch_id;NOT NULL" json:"batch_id"`
	CreatedOn sql.NullTime   `gorm:"column:created_on;type:datetime;comment:创建时间" json:"created_on"`
	CreatedBy sql.NullString `gorm:"column:created_by;type:varchar(64);comment:创建人ID" json:"created_by"`
}

func (m *InstanceLine) TableName() string {
	return "instance_line"
}

// --- 延迟队列 ---

type DelayQueue struct {
	Id          uint64         `gorm:"column:id;type:bigint(20) unsigned;primary_key;AUTO_INCREMENT" json:"id"`
	BizId       string         `gorm:"column:biz_id;type:varchar(64);comment:业务关联ID;NOT NULL" json:"biz_id"`
	ExecuteTime sql.NullTime   `gorm:"column:execute_time;type:datetime(3);comment:期望执行时间;NOT NULL" json:"execute_time"`
	Status      int            `gorm:"column:status;type:tinyint(4);default:0;comment:0待处理,1处理中,2成功,3失败;NOT NULL" json:"status"`
	Payload     sql.NullString `gorm:"column:payload;type:json;comment:消息体" json:"payload"`
	RetryCount  int            `gorm:"column:retry_count;type:int(11);default:0;NOT NULL" json:"retry_count"`
	CreatedAt   sql.NullTime   `gorm:"column:created_at;type:datetime(3);default:CURRENT_TIMESTAMP;NOT NULL" json:"created_at"`
	UpdatedAt   sql.NullTime   `gorm:"column:updated_at;type:datetime(3);default:CURRENT_TIMESTAMP;NOT NULL" json:"updated_at"`
}

func (m *DelayQueue) TableName() string {
	return "delay_queue"
}

type TryNextRow struct {
	DownstreamID string       `gorm:"column:downstream_id"`
	UpstreamID   string       `gorm:"column:upstream_id"`
	UpstreamStat string       `gorm:"column:upstream_status"`
	ExecuteTime  sql.NullTime `gorm:"column:execute_time"`
}
