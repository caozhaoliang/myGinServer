package provider

import (
	"context"
	"io"

	"github.com/google/wire"
)

type Message string

// Options
type Options struct {
	Messages []Message
	Writer   io.Writer
	Reader   io.Reader
}
type Greeter struct {
}

// NewGreeter Greeter的provider方法使用Options以避免构造函数过长
func NewGreeter(ctx context.Context, opts *Options) (*Greeter, error) {
	return nil, nil
}

// GreeterSet 使用wire.Struct设置Options为provider
var GreeterSet = wire.NewSet(wire.Struct(new(Options), "*"), NewGreeter)
