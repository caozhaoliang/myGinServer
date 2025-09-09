package tool

import (
	"io"
	"os"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

func InitLogger() *logrus.Logger {
	path := "./log_myGinServer.log"
	// maxAge := cfg.Log.MaxAge
	/* 日志轮转相关函数
	`WithLinkName` 为最新的日志建立软连接
	`WithRotationTime` 设置日志分割的时间，隔多久分割一次
	WithMaxAge 和 WithRotationCount二者只能设置一个
	  `WithMaxAge` 设置文件清理前的最长保存时间
	  `WithRotationCount` 设置文件清理前最多保存的个数
	*/
	// 下面配置日志每隔 1 分钟轮转一个新文件，保留最近 3 分钟的日志文件，多余的自动清理掉。
	writer, _ := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		// rotatelogs.WithMaxAge(time.Duration(24*maxAge)*time.Hour),
		rotatelogs.WithRotationCount(3),
		rotatelogs.WithRotationTime(time.Duration(24)*time.Hour),
		rotatelogs.WithRotationSize(1024*1024*512),
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.ForceNewFile(),
	)
	output := io.MultiWriter(os.Stdout, writer)
	logger := NewLoggerWithOutput("INFO", output)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "appName"
	}
	fieldHook := NewDefaultFieldHook(map[string]string{
		"app_name": "myGinServer",
		"hostname": hostname,
	})
	logger.AddHook(&fieldHook)

	return logger
}
