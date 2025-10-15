//go:build wireinject
// +build wireinject

package provider

import (
	"context"
	"io"

	"github.com/google/wire"
)

func InitializeGreeter(ctx context.Context, msg []Message, w io.Writer, r io.Reader) (*Greeter, error) {
	wire.Build(GreeterSet)
	return nil, nil
}
