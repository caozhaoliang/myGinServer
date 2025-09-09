package user

type User struct {
	UserId       string `db: user_id`
	Username     string `db: username`
	PasswordHash string `db: password_hash`
	Email        string `db: email`
	Status       string `db: status`
}
