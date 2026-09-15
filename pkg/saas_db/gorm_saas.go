package saas_db

import (
	"bytes"
	"context"
	"strconv"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/pkg/errors"
)

type SaasDb struct {
	conf    *DBConfig
	connMap map[string]*gorm.DB
	dbMutex sync.Mutex
}

func NewSaasDb(conf *DBConfig) *SaasDb {
	return &SaasDb{
		connMap: make(map[string]*gorm.DB),
		conf:    conf,
	}
}

func (d *SaasDb) Close(ctx context.Context) {
	for _, conn := range d.connMap {
		db, err := conn.DB()
		if err == nil {
			_ = db.Close()
		}
	}
}

func (d *SaasDb) GetDB(ctx context.Context, dbName string) (db *gorm.DB, err error) {
	d.dbMutex.Lock()
	defer d.dbMutex.Unlock()
	if len(dbName) == 0 {
		dbName = d.conf.Database
	}
	conn, ok := d.connMap[dbName]
	if ok {
		db = conn.WithContext(ctx)
		return
	}

	var dbCfg = d.conf
	if dbCfg.MaxOpenConns == 0 || dbCfg.MaxIdleConns == 0 {
		dbCfg.MaxOpenConns = 10
		dbCfg.MaxIdleConns = 20
	}

	db, err = CreateEngine(dbName, dbCfg)
	if err != nil {
		return
	}
	d.connMap[dbName] = db
	db = db.WithContext(ctx)
	return
}

func CreateEngine(project string, d *DBConfig) (*gorm.DB, error) {

	var buf bytes.Buffer
	buf.WriteString(d.User)
	buf.WriteString(":")
	buf.WriteString(d.Password)
	buf.WriteString("@tcp(")
	buf.WriteString(d.Host)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(d.Port))
	buf.WriteString(")/")
	if project != "" {
		buf.WriteString(project)
	}
	buf.WriteString("?parseTime=true")
	buf.WriteString("&loc=Local")
	buf.WriteString("&timeout=3s") //设置连接超时时间

	// 连接db实例
	db, err := gorm.Open(mysql.Open(buf.String()), &gorm.Config{})
	if err != nil {
		if db != nil {
			if sqlDB, err := db.DB(); err != nil {
				err = errors.Wrapf(err, "get db connect")
				return nil, err
			} else {
				_ = sqlDB.Close()
			}
		}
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(d.MaxIdleConns)
	sqlDB.SetMaxOpenConns(d.MaxOpenConns)
	return db, nil
}
