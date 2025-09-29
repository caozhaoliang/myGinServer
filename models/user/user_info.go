package user

type User struct {
	UserId       string `db:"user_id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	PhoneNum     string `db:"phone_num"`
	Role         string `db:"role"` // 角色：admin、guest、register、vip、vip2
	Email        string `db:"email"`
	Status       string `db:"status"`
}
