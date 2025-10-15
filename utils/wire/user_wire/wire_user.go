package user_wire

import "github.com/google/wire"

type UserService struct {
	userRepo UserRepository
}

type UserRepository interface {
	GetUserIdID(id int) (any, error)
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		userRepo: repo,
	}
}

type MockUserRepo struct {
	foo string
	bar int
}

func (m *MockUserRepo) GetUserIdID(id int) (any, error) {
	return m.foo, nil
}

// NewMockUserRepo *mockUserRepo构造函数
func NewMockUserRepo(foo string, bar int) *MockUserRepo {
	return &MockUserRepo{
		foo: foo,
		bar: bar,
	}
}

// MockUserRepoSet 将 *mockUserRepo与UserRepository绑定
var MockUserRepoSet = wire.NewSet(NewMockUserRepo, wire.Bind(new(UserRepository), new(*MockUserRepo)))
