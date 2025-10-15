package utils

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/pkg/errors"
)

func TestSlog(t *testing.T) {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelDebug,
	}))
	l.Debug("debug message", "output", "testSlog")

	slog.SetDefault(l)
	slog.Info("info message", "output", "testSlog")

	l.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"info message",
		slog.String("output", "testSlog"),
		slog.Any("err", errors.New("test error")),
	)
	sl := l.With("requestId", "1234567890")
	sl.Debug("with debug")
	l.Info("without others")

	// withGroup
	gsl := l.WithGroup("user").
		With("user_id", "xiaoFen01").
		With("username", "小分")
	gsl.Debug("debug with userName")
	gsl.Info("info withGroup")
}

type User struct {
	Name     string
	Age      int
	Password string
}

func (u *User) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("age", u.Age),
		slog.String("name", u.Name),
	)
}
func TestSlogValue(t *testing.T) {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,           // 记录日志位置
		Level:       slog.LevelDebug, // 设置日志级别
		ReplaceAttr: nil,
	}))

	user := &User{
		Age:      123,
		Name:     "jianghushinian",
		Password: "pass",
	}
	l.Info("info message", "user1", user)

	l.Info("info message", "user2", *user)
}
