package saas_db

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type SaaS interface {
	GetDB(ctx context.Context, project string) (*sqlx.DB, error)
}
type DBConfig struct {
	Database string
	Host     string
	Password string
	Port     int
	User     string

	MaxIdleConns int
	MaxOpenConns int
}
type MysqlSaaSStore struct {
	dbs            map[string]*sqlx.DB
	config         *DBConfig
	dbsManagerLock sync.Mutex
}

func NewSaaS(config *DBConfig) SaaS {
	c := mysql.NewConfig()

	// 2. 设置必需参数
	c.User = config.User                                    // 用户名
	c.Passwd = config.Password                              // 密码（包含特殊字符也无妨）
	c.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port) // 主机:端口
	c.DBName = config.Database                              // 数据库名
	c.Net = "tcp"                                           // 网络协议，默认 tcp

	// 3. 设置连接参数（可选）
	c.Params = map[string]string{
		"charset":   "utf8mb4", // 字符集
		"parseTime": "true",    // 自动处理时间类型
		"loc":       "Local",   // 时区
	}

	// 4. 通过 FormatDSN() 方法生成规范的 DSN 字符串
	dsn := c.FormatDSN()
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		panic(err)
	}
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetMaxOpenConns(config.MaxOpenConns)
	dbs := make(map[string]*sqlx.DB)
	dbs[config.Database] = db
	return &MysqlSaaSStore{dbs, config, sync.Mutex{}}
}

func (s *MysqlSaaSStore) GetDB(ctx context.Context, project string) (*sqlx.DB, error) {
	s.dbsManagerLock.Lock()
	defer s.dbsManagerLock.Unlock()

	if database, ok := s.dbs[project]; ok {
		return database, nil
	}
	config := mysql.NewConfig()

	// 2. 设置必需参数
	config.User = s.config.User                                      // 用户名
	config.Passwd = s.config.Password                                // 密码（包含特殊字符也无妨）
	config.Addr = fmt.Sprintf("%s:%d", s.config.Host, s.config.Port) // 主机:端口
	config.DBName = project                                          // 数据库名
	config.Net = "tcp"                                               // 网络协议，默认 tcp

	// 3. 设置连接参数（可选）
	config.Params = map[string]string{
		"charset":   "utf8mb4", // 字符集
		"parseTime": "true",    // 自动处理时间类型
		"loc":       "Local",   // 时区
	}

	// 4. 通过 FormatDSN() 方法生成规范的 DSN 字符串
	dsn := config.FormatDSN()
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		err = errors.Wrapf(err, "failed to open project database: %s", project)
	} else {
		db.SetMaxIdleConns(s.config.MaxIdleConns)
		db.SetMaxOpenConns(s.config.MaxOpenConns)
		s.dbs[project] = db
	}
	return db, err
}
