package store

import (
	"context"
	"database/sql"
	"fmt"
	"myGinServer/config"
	"myGinServer/models/article"
	"myGinServer/models/task"
	"myGinServer/models/tenant"
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
	ListUsers(ctx context.Context, keyword string, page, size int) ([]user2.User, int64, error)
	UpdateUser(ctx context.Context, user user2.User) error
	UpdateUserStatus(ctx context.Context, id, status string) error
	UpdateUserPassword(ctx context.Context, id, hashed string) error
	GetUserWithPassword(ctx context.Context, id string) (user2.User, error)
	UpdateUserProfile(ctx context.Context, id, nickname, email, phoneNum string) error

	ListUserTenants(ctx context.Context, userId string) ([]tenant.TenantMember, error)
	GetTenantMember(ctx context.Context, userId, tenantId string) (tenant.Tenant, error)
	EnsureDefaultTenant(ctx context.Context) error

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
	sqlText := `select user_id, username, nickname, avatar, email, phone_num, role, status from users where username=?`
	err = s.db.SelectContext(ctx, &user, sqlText, username)
	if errors.Is(err, sql.ErrNoRows) {
		return user2.User{}, nil
	}
	return user, err
}
func (s *DbStore) GetUser(ctx context.Context, id string) (user user2.User, err error) {
	sqlText := `select user_id, username, nickname, avatar, email, phone_num, role, status from users where user_id=? limit 1`
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
	sqlText := `INSERT INTO users(user_id, username, email, phone_num, nickname, avatar, role, password_hash, status) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, sqlText, user.UserId, user.Username, user.Email,
		user.PhoneNum, user.Nickname, user.Avatar, user.Role, user.PasswordHash, user.Status)
	return err
}

// ListUsers 分页查询用户列表，keyword 为空时返回全部。
func (s *DbStore) ListUsers(ctx context.Context, keyword string, page, size int) ([]user2.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	where := ""
	args := []interface{}{}
	if keyword != "" {
		where = " WHERE username LIKE concat('%', ?, '%') OR nickname LIKE concat('%', ?, '%') OR email LIKE concat('%', ?, '%')"
		args = append(args, keyword, keyword, keyword)
	}
	var total int64
	countSQL := `SELECT COUNT(*) FROM users` + where
	if err := s.db.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	listSQL := `SELECT user_id, username, nickname, avatar, email, phone_num, role, status FROM users` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, size, (page-1)*size)
	var users []user2.User
	if err := s.db.SelectContext(ctx, &users, listSQL, args...); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// UpdateUser 更新用户基础信息（邮箱/手机号/昵称/头像/角色），不含密码与状态。
func (s *DbStore) UpdateUser(ctx context.Context, user user2.User) error {
	sqlText := `UPDATE users SET email=?, phone_num=?, nickname=?, avatar=?, role=? WHERE user_id=?`
	_, err := s.db.ExecContext(ctx, sqlText, user.Email, user.PhoneNum, user.Nickname, user.Avatar, user.Role, user.UserId)
	return err
}

// UpdateUserStatus 更新用户账号状态（如禁用）。
func (s *DbStore) UpdateUserStatus(ctx context.Context, id, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET status=? WHERE user_id=?`, status, id)
	return err
}

// UpdateUserPassword 更新用户密码哈希（重置密码 / 修改密码）。
func (s *DbStore) UpdateUserPassword(ctx context.Context, id, hashed string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE user_id=?`, hashed, id)
	return err
}

// GetUserWithPassword 获取包含密码哈希的完整用户信息，用于修改密码时比对旧密码。
func (s *DbStore) GetUserWithPassword(ctx context.Context, id string) (user2.User, error) {
	sqlText := `SELECT user_id, username, password_hash, nickname, avatar, email, phone_num, role, status FROM users WHERE user_id=? LIMIT 1`
	var user user2.User
	err := s.db.GetContext(ctx, &user, sqlText, id)
	return user, err
}

// UpdateUserProfile 更新用户个人资料（昵称/邮箱/手机号）。
func (s *DbStore) UpdateUserProfile(ctx context.Context, id, nickname, email, phoneNum string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE users SET nickname=?, email=?, phone_num=? WHERE user_id=?`, nickname, email, phoneNum, id)
	return err
}

// ListUserTenants 返回用户参与的所有租户及其角色。
func (s *DbStore) ListUserTenants(ctx context.Context, userId string) ([]tenant.TenantMember, error) {
	sqlText := `SELECT t.tenant_id, t.name, t.code, t.database_name, t.status, ut.role_in_tenant, ut.is_default
		FROM user_tenant ut JOIN tenant t ON ut.tenant_id = t.tenant_id
		WHERE ut.user_id = ? AND ut.status = 'active' ORDER BY ut.is_default DESC, ut.created_on ASC`
	var members []tenant.TenantMember
	err := s.db.SelectContext(ctx, &members, sqlText, userId)
	return members, err
}

// GetTenantMember 校验用户是否为指定租户成员，并返回租户信息（含 database_name）。
func (s *DbStore) GetTenantMember(ctx context.Context, userId, tenantId string) (tenant.Tenant, error) {
	sqlText := `SELECT t.tenant_id, t.name, t.code, t.database_name, t.status, t.created_on, t.created_by
		FROM user_tenant ut JOIN tenant t ON ut.tenant_id = t.tenant_id
		WHERE ut.user_id = ? AND ut.tenant_id = ? AND ut.status = 'active' LIMIT 1`
	var t tenant.Tenant
	err := s.db.GetContext(ctx, &t, sqlText, userId, tenantId)
	return t, err
}

// EnsureDefaultTenant 幂等初始化默认租户，并把所有 admin 用户以 owner 身份绑定到默认租户。
func (s *DbStore) EnsureDefaultTenant(ctx context.Context) error {
	// 1. 默认租户
	_, err := s.db.ExecContext(ctx, `INSERT IGNORE INTO tenant(tenant_id, name, code, database_name, status, created_on, created_by)
		VALUES('default', '默认租户', 'default', 'mytest', 'active', NOW(), 'system')`)
	if err != nil {
		return err
	}
	// 2. 绑定所有 admin 用户为 owner（靠 UNIQUE(user_id, tenant_id) 保证幂等）
	_, err = s.db.ExecContext(ctx, `INSERT IGNORE INTO user_tenant(user_id, tenant_id, role_in_tenant, is_default, status, created_on)
		SELECT user_id, 'default', 'owner', 1, 'active', NOW() FROM users WHERE role = 'admin'`)
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
