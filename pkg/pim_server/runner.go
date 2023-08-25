package pim_server

import (
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net"
	"os"
	"os/signal"
	"pim/api"
	"pim/pkg/models"
	"pim/pkg/pim_server/dao"
	"pim/pkg/promePkg"
	"pim/pkg/tools"
	"sync"
	"syscall"
	"time"
)

type server struct {
	logger          *zap.Logger
	redisPool       *redis.Pool
	db              *gorm.DB
	closeServer     chan struct{}
	config          *ImServerConfig
	mu              sync.RWMutex
	dao             dao.APIDao
	dbLogger        *DBLogger
	grpcd           *grpc.Server
	pim             *PimServer
	rpc_port        int
	msgNode         *tools.Node
	sendMessageChan chan *api.Message
	saveMessageChan chan *models.SingleMessage
}

type DBLogger struct {
	logger *zap.Logger
}

func (D DBLogger) Printf(s string, i ...interface{}) {
	//TODO implement me
	//panic("implement me")
	D.logger.Debug("DBLog", zap.String("log", fmt.Sprintf(s, i...)))
}

type Option func(svr *server)

func SetRedis(url, password string, db int) Option {
	return func(svr *server) {
		var err error
		svr.redisPool, err = tools.NewRedis(url, password, db)

		if err != nil {
			panic(err)
		}
	}
}

func (s *server) GetWriter() logger.Writer {
	return s.dbLogger
}

func (s *server) Run() error {

	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it

	// 开启端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.rpc_port))
	if err != nil {
		fmt.Printf("开启服务失败: %s", err)
		return nil
	}

	go s.StartSenderMessageEventService()
	go s.StartSaveMessageEventService()
	go s.StartAutoClearFileTaskService()

	go func() {
		// service connections
		err = s.grpcd.Serve(lis)
		if err != nil {
			fmt.Printf("开启服务失败: %s", err)
		}
		close(quit)
		//if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		//	log.Fatalf("listen: %s\n", err)
		//}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 等待终止信号
	<-quit
	close(s.closeServer)
	//log.Println("Shutdown Server ...")
	s.logger.Error("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	//if err := srv.Shutdown(ctx); err != nil {
	//	s.logger.Error("Server Shutdown err :", zap.Error(err))
	//}
	close(s.closeServer)
	s.grpcd.GracefulStop()

	select {
	case <-ctx.Done():
		s.logger.Info("timeout of 1 seconds.")
	}
	s.logger.Error("Server exiting")
	return nil
}
func SetMysqlDBByConfig(dbUri string) Option {
	return func(svr *server) {
		db, err := gorm.Open(mysql.New(mysql.Config{
			DSN: dbUri,
			//DefaultStringSize: 256, // string 类型字段的默认长度
			//DisableDatetimePrecision: true, // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
			//DontSupportRenameIndex: true, // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
			//DontSupportRenameColumn: true, // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
			SkipInitializeWithVersion: true, // 根据当前 MySQL 版本自动配置
		}), &gorm.Config{
			Logger: logger.New(svr.GetWriter(), logger.Config{
				LogLevel: logger.Info,
			}),
		})
		if err != nil {
			panic(err)
		}

		svr.db = db
		mgTool := db.Set("gorm:table_options", "ENGINE = InnoDB DEFAULT CHARSET=utf8mb4").Migrator()

		_ = mgTool
		mgTool.AutoMigrate(&models.SingleMessage{})

	}
}

// RunApp 启动服务
func RunApp(config *ImServerConfig) {
	//	初始化日志系统
	configLevel, err := zapcore.ParseLevel(config.LoggerLevel)
	// 未设置默认warn
	if err != nil {
		configLevel = zapcore.InfoLevel
	}

	logger := tools.LoggerInitLevelTag(config.LoggerPath, "pim", &configLevel)

	//logger, _ := zap.NewDevelopment()
	logger.Error("启动参数",
		zap.Any("params", config),
	)

	//
	svr := &server{
		logger:      logger,
		config:      config,
		closeServer: make(chan struct{}),
	}
	// 数据库日志
	svr.dbLogger = &DBLogger{
		logger: svr.logger,
	}

	SetNodeID()(svr)
	SetMessageChan()(svr)
	// redis池
	SetRedis(config.RedisIP, config.RedisPassword, config.RedisDB)(svr)

	// 设置数据库
	SetMysqlDBByConfig(config.DBUri)(svr)

	// dao
	svr.dao = dao.NewDao(svr.logger, svr.db, svr.redisPool, svr.closeServer)

	// grpc
	SetRpcService(config.RpcPort)(svr)

	// 计数器
	promePkg.InitConter()

	svr.Run()

}

func SetMessageChan() Option {
	return func(svr *server) {
		svr.sendMessageChan = make(chan *api.Message, 16)
		svr.saveMessageChan = make(chan *models.SingleMessage, 16)
	}
}

// SetRpcService 注册pim服务
func SetRpcService(port int) Option {
	return func(svr *server) {
		svr.rpc_port = port
		svr.pim = &PimServer{
			svr:                 svr,
			rw:                  new(sync.RWMutex),
			clients:             make(map[int64]*RpcClient, 128),
			UserStreamClientMap: make(UserStreamClientMapType, 8),
			//GroupsCache:         make(GroupsCache, 8),
		}

		svr.grpcd = grpc.NewServer()

		// 参数分别是grpc服务与自己的服务
		api.RegisterPimServerServer(svr.grpcd, svr.pim)
	}
}
