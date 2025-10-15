//go:build wireinject
// +build wireinject

package wire

import "github.com/google/wire"

var EventSet = wire.NewSet(NewEvent, NewGreeter, NewMessage)

func InitializeEvent() Event {
	wire.Build(EventSet)
	return Event{}
}
