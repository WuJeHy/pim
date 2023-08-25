package dao

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"pim/pkg/tools"
	"testing"
	"time"
)

func NewDemoDaoClient(closeStatus chan struct{}) APIDao {

	logger := zap.NewExample()
	dbUri := "todo" // 这里需要改成自己的
	redisServer := "127.0.0.1"
	redisPassword := ""
	redisDB := 8
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dbUri,
		//DefaultStringSize: 256, // string 类型字段的默认长度
		//DisableDatetimePrecision: true, // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		//DontSupportRenameIndex: true, // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		//DontSupportRenameColumn: true, // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: true, // 根据当前 MySQL 版本自动配置
	}))
	if err != nil {
		panic(err)
	}

	redis, err := tools.NewRedis(redisServer, redisPassword, redisDB)
	if err != nil {
		panic(err)
	}
	daoClient := NewDao(logger, db, redis, closeStatus)

	return daoClient
}

var testDBClient APIDao
var closeStatus chan struct{}

func initTest() {
	if testDBClient == nil {

		closeStatus = make(chan struct{})
		testDBClient = NewDemoDaoClient(closeStatus)

		go testDBClient.AutoClearService()
	}
}

func deinitTest() {
	close(closeStatus)
}

func TestKV(t *testing.T) {

	initTest()
	timeA := time.Now().UnixNano()
	AData, _ := testDBClient.GetChatInfoByID(2, 2)
	timeB := time.Now().UnixNano()
	BData, _ := testDBClient.GetChatInfoByID(2, 2)
	timeC := time.Now().UnixNano()

	_ = AData
	_ = BData
	fmt.Println("FT:", timeB-timeA, " LT:", timeC-timeB)

	deinitTest()
}
