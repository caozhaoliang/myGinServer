package store

import (
	"context"
	"database/sql"
	"fmt"
	"myGinServer/config"
	"myGinServer/models/article"
	"myGinServer/models/task"
	user2 "myGinServer/models/user"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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
	GetUserByName(ctx context.Context, username string) (user user2.User, err error)
	GetUser(ctx context.Context, id string) (user user2.User, err error)

	ListTasks(ctx context.Context, keyword string) ([]task.Task, error)
	DelTasks(ctx context.Context, id string) error
	SaveTask(ctx context.Context, task task.Task) (string, error)

	Channels(ctx context.Context) ([]article.Channel, error)
	DeleteArticle(ctx context.Context, id string) error
	SaveArticle(ctx context.Context, article *article.ArticleVO) error
	GetArticle(ctx context.Context, id string) (article *article.ArticleVO, err error)
	UpdateArticle(ctx context.Context, article *article.ArticleVO) error
	GetArticles(ctx context.Context, req *article.ArticlesRequest) (article.ArticlesResponse, error)
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

func (s *DbStore) GetUserByName(ctx context.Context, username string) (user user2.User, err error) {
	sqlText := `select user_id, username, email, phone_num, role, status from users where username=?`
	err = s.db.SelectContext(ctx, &user, sqlText, username)
	if errors.Is(err, sql.ErrNoRows) {
		return user2.User{}, nil
	}
	return user, err
}
func (s *DbStore) GetUser(ctx context.Context, id string) (user user2.User, err error) {
	sqlText := `select user_id, username, email, phone_num, role, status from users where user_id=? limit 1`
	err = s.db.GetContext(ctx, &user, sqlText, id)
	if errors.Is(err, sql.ErrNoRows) {
		return user2.User{}, nil
	}
	return user, err
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
	sqlText := `INSERT INTO users(user_id, username, email, role, password_hash, status) VALUES(?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, sqlText, user.UserId, user.Username, user.Email,
		user.Role, user.PasswordHash, user.Status)
	return err
}

func (s *DbStore) ListTasks(ctx context.Context, keyword string) ([]task.Task, error) {
	var tasks []task.Task
	arg := []interface{}{}
	query := `SELECT id, name, des FROM tasks`
	if keyword != "" {
		arg = append(arg, keyword)
		query += ` WHERE name LIKE concat('%', ?, '%')`
	}

	err := s.db.SelectContext(ctx, &tasks, query, arg...)
	return tasks, err
}

func (s *DbStore) DelTasks(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func (s *DbStore) SaveTask(ctx context.Context, task task.Task) (string, error) {
	execSQL := `insert into tasks(id, name, des, completed) values (?,?,?,?)`
	if len(task.Id) == 0 {
		newUUID, _ := uuid.NewUUID()
		task.Id = newUUID.String()
	}
	_, err := s.db.ExecContext(ctx, execSQL, task.Id, task.Name, task.Des, task.Completed)
	return task.Id, err
}

func (s *DbStore) Channels(ctx context.Context) ([]article.Channel, error) {
	sqlText := `select id, name, des, status from channel`
	var channels []article.Channel
	err := s.db.SelectContext(ctx, &channels, sqlText)
	return channels, err
}

func (s *DbStore) DeleteArticle(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM articles WHERE id = ?`, id)
	return err
}

func (s *DbStore) SaveArticle(ctx context.Context, article *article.ArticleVO) error {
	execSQL := `insert into articles(id, title,content, cover, channel_id, 
                     status, pubdate, view_count, like_count, comment_count) 
				values (?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.ExecContext(ctx, execSQL, article.Id, article.Title, article.Content, article.Cover, article.ChannelId,
		article.Status, article.Pubdate, article.ViewCount, article.LikeCount, article.CommentCount)
	return err
}
func (s *DbStore) GetArticle(ctx context.Context, id string) (*article.ArticleVO, error) {
	sqlText := `select id, title, content, cover, channel_id, status, pubdate, view_count
     , like_count, comment_count from articles where id = ?`
	articleVo := new(article.ArticleVO)
	err := s.db.GetContext(ctx, articleVo, sqlText, id)
	return articleVo, err
}

func (s *DbStore) UpdateArticle(ctx context.Context, article *article.ArticleVO) error {
	execSQL := `update articles set title = ?,content=?, cover = ?, channel_id = ?
				where id = ?`
	_, err := s.db.ExecContext(ctx, execSQL, article.Title, article.Content,
		article.Cover, article.ChannelId, article.Id)
	return err
}

func (s *DbStore) GetArticles(ctx context.Context, req *article.ArticlesRequest) (article.ArticlesResponse, error) {
	var response article.ArticlesResponse
	query := `select id, title,content, cover,channel_id, status, pubdate, view_count, like_count, comment_count from articles`
	countQuery := `SELECT COUNT(*) FROM articles`

	// 构建条件部分
	var args []interface{}
	var conditions []string

	// 状态过滤
	if req.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, req.Status)
	}

	// 频道ID过滤
	if req.ChannelId != "" {
		conditions = append(conditions, "channel_id = ?")
		args = append(args, req.ChannelId)
	}

	// 发布日期范围过滤
	if req.PubdateStart != "" {
		conditions = append(conditions, "pubdate >= ?")
		args = append(args, req.PubdateStart)
	}

	if req.PubdateEnd != "" {
		conditions = append(conditions, "pubdate <= ?")
		args = append(args, req.PubdateEnd)
	}

	// 添加WHERE条件到查询语句
	if len(conditions) > 0 {
		whereClause := " WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
		query += whereClause
		countQuery += whereClause
	}

	// 获取总数
	err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&response.Total)
	if err != nil {
		return response, err
	}

	// 添加分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	query += " ORDER BY pubdate DESC LIMIT ? OFFSET ?"
	args = append(args, req.PageSize, offset)

	// 执行查询
	err = s.db.SelectContext(ctx, &response.Data, query, args...)
	if err != nil {
		return response, err
	}

	return response, nil
}
