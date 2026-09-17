package user

type User struct {
	UserId       string `db:"user_id" json:"user_id"`
	Username     string `db:"username" json:"username"`
	PasswordHash string `db:"password_hash" json:"-"`
	Nickname     string `db:"nickname" json:"nickname"`
	Avatar       string `db:"avatar" json:"avatar"`
	PhoneNum     string `db:"phone_num" json:"phone_num"`
	Role         string `db:"role" json:"role"` // 角色：admin、guest、register、vip、vip2
	Email        string `db:"email" json:"email"`
	Status       string `db:"status" json:"status"`
}
