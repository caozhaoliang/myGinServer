package store

import (
	"context"
	"database/sql"
	"fmt"
	"myGinServer/config"
	user2 "myGinServer/internal/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type DbStore struct {
	db *sqlx.DB
}

type DBStore interface {
	CheckUserInfo(ctx context.Context, username, password string) (user2.User, error)
	IsExistsUserNameEmail(ctx context.Context, username, email string) (bool, error)
	SaveUser(ctx context.Context, user user2.User) error
}

func NewDatabase(config *config.DBConfig) (DBStore, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local",
		config.DbUser, config.DbPassword, config.DbHost, config.DbPort, config.DbName)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &DbStore{db: db}, nil
}

func (s *DbStore) IsExistsUserNameEmail(ctx context.Context, username, email string) (bool, error) {
	sqlText := `SELECT username, email FROM users WHERE username = ? OR email = ?`
	var usernameExists, emailExists string
	err := s.db.QueryRowxContext(ctx, sqlText, username, email).Scan(&usernameExists, &emailExists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return usernameExists != "" || emailExists != "", nil
}

func (s *DbStore) CheckUserInfo(ctx context.Context, username, password string) (user2.User, error) {

	var user user2.User
	err := s.db.QueryRowxContext(ctx, `SELECT user_id, username, password_hash, status 
            FROM users 
            WHERE username = ? limit 1`, username).Scan(
		&user.UserId, &user.Username, &user.PasswordHash, &user.Status,
	)
	if err != nil {
		return user, errors.New("用户名或密码错误")
	}
	if user.Status != "active" {
		return user, errors.New("账号已禁用，请联系管理员")
	}

	// 4. 验证密码（对比哈希值，避免明文存储）
	if err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return user, errors.New("用户名或密码错误") // 密码错误
	}
	return user, nil
}

func (s *DbStore) SaveUser(ctx context.Context, user user2.User) error {
	sqlText := `INSERT INTO users(user_id, username, email, password_hash, status) VALUES(?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, sqlText, user.UserId, user.Username, user.Email, user.PasswordHash, user.Status)
	return err
}
