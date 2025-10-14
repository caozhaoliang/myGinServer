package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GoSystemInfo struct {
	ID         int    `gorm:"primary_key"`
	SystemId   string `gorm:"column:systemId;type:varchar(30);not null;index:SystemId"`
	SystemName string `gorm:"column:systemName;type:varchar(50);not null;default:'defaultNull'"`
}

func (t *GoSystemInfo) TableName() string {
	return "go_system_info"
}

type GoServiceInfo struct {
	//gorm.Model
	ID          uint   `gorm:"primary_key"`
	SystemId    string `gorm:"column:systemId;type:varchar(30);not null;index:SystemId"`
	ServiceId   string `gorm:"column:serviceId;type:varchar(50);not null;default:'defaultNull';index:ServiceId"`
	ServiceName string `gorm:"column:serviceName;type:varchar(50);not null;default:'defaultNull'"`
}

func (t *GoServiceInfo) TableName() string {
	return "go_service_info"
}

type result struct {
	SystemId    string `json:"systemId"`
	SystemName  string `json:"systemName"`
	ServiceId   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
}

// 定义数据库连接
type ConnInfo struct {
	MyUser   string
	Password string
	Host     string
	Port     int
	Db       string
}

func dbConn(MyUser, Password, Host, Db string, Port int) *gorm.DB {
	connArgs := fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", MyUser, Password, Host, Port, Db)
	dbHandler, err := gorm.Open(mysql.Open(connArgs), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
	return dbHandler
}
func mapToJson(result interface{}) string {
	// map转 json str
	jsonBytes, _ := json.Marshal(result)
	jsonStr := string(jsonBytes)
	return jsonStr
}

var db *gorm.DB

func InitDB() {
	once := sync.Once{}
	once.Do(func() {
		cn := ConnInfo{
			"root",
			"12345678",
			"localhost",
			3306,
			"grom",
		}
		db = dbConn(cn.MyUser, cn.Password, cn.Host, cn.Db, cn.Port)
	})
}

func TestInit(t *testing.T) {
	InitDB()
	// 创建表
	db.AutoMigrate(&GoSystemInfo{})
	product := GoSystemInfo{SystemId: "sysid", SystemName: "sysname"}
	fmt.Println(product)
	db.AutoMigrate(&GoServiceInfo{})
	products := GoServiceInfo{SystemId: "sysid", ServiceId: "serid", ServiceName: "sername"}
	fmt.Println(products)

}

func TestInsert(t *testing.T) {
	InitDB()
	//var serverInfo = &GoServiceInfo{
	//	ServiceId:   "myServiceId001",
	//	ServiceName: "myServiceName001",
	//	SystemId:    "mySystemId001",
	//}
	var systemInfo = &GoSystemInfo{
		SystemId:   "mySystemId001",
		SystemName: "mySystemName001",
	}
	err := db.Save(systemInfo).Error
	assert.Nil(t, err)
}

func TestQuery(t *testing.T) {
	InitDB()
	var results []result
	db.Table("go_service_info").Select(`go_service_info.serviceId as service_id, go_service_info.serviceName as service_name,
go_system_info.systemId as system_id, go_system_info.systemName as system_name`).
		Joins(`left join go_system_info on go_service_info.systemId = go_system_info.systemId`).Scan(&results)
	fmt.Println(mapToJson(results))

	db.Table("go_service_info").Select(`go_service_info.serviceId as service_id, go_service_info.serviceName as service_name,
go_system_info.systemId as system_id, go_system_info.systemName as system_name`).
		Joins(`left join go_system_info on go_service_info.systemId = go_system_info.systemId 
where go_service_info.serviceId <> ? and go_system_info.systemId = ?`, "xxx", "mySystemId001").Scan(&results)
	fmt.Println(mapToJson(results))

	// 原生sql
	db.Raw(`SELECT a.serviceId as service_id,a.serviceName as service_name, b.systemId as system_id,
b.systemName as system_name FROM go_service_info a LEFT JOIN go_system_info b ON a.systemId = b.systemId`).Scan(&results)
	fmt.Println(mapToJson(results))
	// where
	db.Raw(`SELECT a.serviceId as service_id,a.serviceName as service_name, b.systemId as system_id, 
b.systemName as system_name FROM go_service_info a LEFT JOIN go_system_info b ON a.systemId = b.systemId 
                            where a.serviceId <> ? and b.systemId = ?`, "xxx", "mySystemId001").Scan(&results)
	fmt.Println(mapToJson(results))
}

func TestUpdate(t *testing.T) {
	InitDB()
	err := db.Table("go_service_info").
		Where("serviceId = ?", "myServiceId001").
		Update("serviceName", "myServiceName0011").Error
	assert.Nil(t, err)
}

func TestInsertMany(t *testing.T) {
	InitDB()
	var serviceInfos = []GoServiceInfo{
		{SystemId: "sysid001", ServiceId: "serid001", ServiceName: "sername001"},
		{SystemId: "sysid002", ServiceId: "serid002", ServiceName: "sername002"},
		{SystemId: "sysid003", ServiceId: "serid003", ServiceName: "sername003"},
	}

	db.CreateInBatches(&serviceInfos, 10)
	fmt.Println(serviceInfos)
}

func TestDelete(t *testing.T) {
	InitDB()
	err := db.Table("go_service_info").
		Where("serviceId = ?", "serid003").
		Delete(GoServiceInfo{}).Error
	assert.Nil(t, err)
}

func TestList(t *testing.T) {
	InitDB()
	services := []GoServiceInfo{}
	err := db.Find(&services).Error
	assert.Nil(t, err)
	fmt.Println(services)
}
