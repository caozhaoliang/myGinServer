//go:build wireinject
// +build wireinject

package user_wire

import "github.com/google/wire"

func InitializeUserService(foo string, bar int) *UserService {
	wire.Build(NewUserService, MockUserRepoSet)
	return nil
}
