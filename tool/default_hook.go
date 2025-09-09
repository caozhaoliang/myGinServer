package tool

import (
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

type DefaultFieldHook struct {
	fields map[string]string
}

func (h *DefaultFieldHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *DefaultFieldHook) Fire(e *logrus.Entry) error {
	for k, v := range h.fields {
		e.Data[k] = v
	}
	return nil
}

func NewDefaultFieldHook(fields map[string]string) DefaultFieldHook {
	return DefaultFieldHook{
		fields: fields,
	}
}

func NewLoggerWithOutput(levelName string, output io.Writer, hooks ...logrus.Hook) *logrus.Logger {
	level, err := logrus.ParseLevel(levelName)
	if err != nil {
		level = logrus.ErrorLevel
	}

	logFormatter := &logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		FullTimestamp:   true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcname := s[len(s)-1]
			_, filename := path.Split(f.File)
			return funcname, fmt.Sprintf("%s:%v", filename, f.Line)
		},
	}

	defaultFieldsHook := NewDefaultFieldHook(map[string]string{
		"env_code": "local",
	})
	logger := logrus.New()
	logger.AddHook(&defaultFieldsHook)
	logger.SetLevel(level)
	logger.SetOutput(output)
	logger.SetFormatter(logFormatter)
	logger.SetReportCaller(true)

	for _, hook := range hooks {
		logger.AddHook(hook)
	}

	// 设置全局log. 一些第三方库使用到全局log
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(logFormatter)
	logrus.SetReportCaller(true)
	for _, hook := range hooks {
		logrus.AddHook(hook)
	}

	return logger
}
