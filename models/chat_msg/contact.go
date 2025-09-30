package chat_msg

// Contact 关系表
type Contact struct {
	Id       int64  `db:"id"`
	OwnerId  string `db:"owner_id"`  // 关系的拥有者
	TargetId string `db:"target_id"` // 关系对象
	Type     string `db:"type"`      // 类型：person(个人),group(群组)
	Desc     string `db:"desc"`
}
