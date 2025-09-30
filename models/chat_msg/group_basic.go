package chat_msg

type GroupBasic struct {
	Id      int64  `db:"id"`
	Name    string `db:"name"`
	OwnerId string `db:"owner_id"`
	Icon    string `db:"icon"`
	Type    string `db:"type"`
	Desc    string `db:"desc"`
}
