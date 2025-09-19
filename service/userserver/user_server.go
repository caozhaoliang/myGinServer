package userserver

import (
	"context"
	"myGinServer/internal/store"
	user2 "myGinServer/models/user"

	"github.com/google/uuid"
	"github.com/pkg/errors"
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
