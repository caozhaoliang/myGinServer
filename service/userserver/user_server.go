package userserver

import (
	"context"
	"myGinServer/api/request"
	"myGinServer/internal/store"
	user2 "myGinServer/models/user"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type UserServer struct {
	dbStore store.DBStore
}

type UserManagerService interface {
	// Register(ctx context.Context, user *user2.User) (string, error)
}

func NewUserServer(dbStore store.DBStore) *UserServer {
	return &UserServer{
		dbStore: dbStore,
	}
}

var (
	errUserExists      = errors.New("用户已存在")
	errInvalidPassword = errors.New("密码格式错误")
	errEmailExists     = errors.New("邮箱或者用户名已存在")
	errInvalidEmail    = errors.New("邮箱格式错误")
	errPasswordHash    = errors.New("密码加密错误")
	errRegisterFailed  = errors.New("注册失败")
)

func (u *UserServer) Register(user *user2.User) (userId string, err error) {

	// 额外的密码复杂度验证
	if err = validatePassword(user.PasswordHash); err != nil {
		return "", err
	}

	exists, err := u.dbStore.IsExistsUserNameEmail(context.Background(), user.Username, user.Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errEmailExists
	}

	// 密码加密
	hashedPassword, err := hashPassword(user.PasswordHash)
	if err != nil {
		return "", errors.Wrap(errPasswordHash, err.Error())
	}
	userVO := user2.User{
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: hashedPassword,
		Role:         "register",
		Status:       "active",
		UserId:       uuid.New().String(),
	}

	// 保存用户到数据库
	if err = u.dbStore.SaveUser(context.Background(), userVO); err != nil {
		return "", errors.Wrap(errRegisterFailed, err.Error())
	}

	return userVO.UserId, nil
}

func (u *UserServer) Profile(ctx context.Context, userId string) (*user2.User, error) {
	user, err := u.dbStore.GetUser(ctx, userId)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile 更新当前用户个人资料（昵称/邮箱/手机号）。
func (u *UserServer) UpdateProfile(ctx context.Context, userId string, req *request.UpdateProfileReq) error {
	return u.dbStore.UpdateUserProfile(ctx, userId, req.Nickname, req.Email, req.PhoneNum)
}

// ChangePassword 修改当前用户密码：校验旧密码 + 校验新密码复杂度。
func (u *UserServer) ChangePassword(ctx context.Context, userId string, req *request.ChangePasswordReq) error {
	user, err := u.dbStore.GetUserWithPassword(ctx, userId)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errors.New("旧密码错误")
	}
	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}
	hashed, err := hashPassword(req.NewPassword)
	if err != nil {
		return errors.Wrap(errPasswordHash, err.Error())
	}
	return u.dbStore.UpdateUserPassword(ctx, userId, hashed)
}

// List 分页查询用户列表。
func (u *UserServer) List(ctx context.Context, keyword string, page, size int) ([]user2.User, int64, error) {
	return u.dbStore.ListUsers(ctx, keyword, page, size)
}

// Create 管理员创建用户：校验用户名/邮箱唯一，加密密码后落库。
func (u *UserServer) Create(ctx context.Context, req *request.CreateUserReq) (string, error) {
	if err := validatePassword(req.Password); err != nil {
		return "", err
	}
	exists, err := u.dbStore.IsExistsUserNameEmail(ctx, req.Username, req.Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errEmailExists
	}
	hashed, err := hashPassword(req.Password)
	if err != nil {
		return "", errors.Wrap(errPasswordHash, err.Error())
	}
	role := req.Role
	if role == "" {
		role = "register"
	}
	userVO := user2.User{
		UserId:       uuid.New().String(),
		Username:     req.Username,
		Email:        req.Email,
		PhoneNum:     req.PhoneNum,
		Nickname:     req.Nickname,
		Role:         role,
		Status:       "active",
		PasswordHash: hashed,
	}
	if err := u.dbStore.SaveUser(ctx, userVO); err != nil {
		return "", errors.Wrap(errRegisterFailed, err.Error())
	}
	return userVO.UserId, nil
}

// Update 管理员更新用户基础信息（不含密码与状态）。
func (u *UserServer) Update(ctx context.Context, id string, req *request.UpdateUserReq) error {
	return u.dbStore.UpdateUser(ctx, user2.User{
		UserId:   id,
		Email:    req.Email,
		PhoneNum: req.PhoneNum,
		Nickname: req.Nickname,
		Role:     req.Role,
	})
}

// Disable 逻辑禁用用户，禁止禁用自己。
func (u *UserServer) Disable(ctx context.Context, id, operatorId string) error {
	if id == operatorId {
		return errors.New("不能禁用自己")
	}
	return u.dbStore.UpdateUserStatus(ctx, id, "disabled")
}

// ResetPassword 管理员重置用户密码。
func (u *UserServer) ResetPassword(ctx context.Context, id string, newPwd string) error {
	if err := validatePassword(newPwd); err != nil {
		return err
	}
	hashed, err := hashPassword(newPwd)
	if err != nil {
		return errors.Wrap(errPasswordHash, err.Error())
	}
	return u.dbStore.UpdateUserPassword(ctx, id, hashed)
}
