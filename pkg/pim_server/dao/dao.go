package dao

import (
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"pim/api"
	"pim/pkg/codes"
	"pim/pkg/models"
	"pim/pkg/tools"
	"sync"
	"time"
)

type CacheValue struct {
	UpdateAt int64       // UpdateAt 最后一次取的时间
	Value    interface{} // Value 存实际的对象指针
}

func (c *CacheValue) UpdateTime() {
	c.UpdateAt = time.Now().Unix()
}

func (c *CacheValue) UpdateValue(value interface{}) {
	c.UpdateAt = time.Now().Unix()
	c.Value = value
}

func NewValue(value interface{}) *CacheValue {
	return &CacheValue{
		UpdateAt: time.Now().Unix(),
		Value:    value,
	}
}

type Dao struct {
	closeServer <-chan struct{}

	logger         *zap.Logger
	db             *gorm.DB
	redisPool      *redis.Pool
	rw             *sync.RWMutex
	cacheGlobalMap map[string]*CacheValue
}

func GetChatInfoCacheValue(d *Dao, key string) (info *api.ChatInfoDataType, err error) {
	// 同时更新有效期
	cacheValue, err := d.GetCacheKey(key, true)

	if err != nil {
		return
	}

	switch value := cacheValue.(type) {
	case *api.ChatInfoDataType:
		return value, nil
	default:
		return nil, errors.New("缓存的类型对不上")
	}
}

// GetChatInfoByID 获取与某人的聊天记录（这个是指A是否与B聊过天）
func (d *Dao) GetChatInfoByID(myUserID int64, chatID int64) (info *api.ChatInfoDataType, err error) {
	//TODO implement me
	//panic("implement me")

	// 检查用户ID是否合法
	if myUserID <= 0 {
		err = errors.New("userid fail")
		return
	}

	// 从redis池中获取一个连接
	redisConn := d.redisPool.Get()
	// 最后关闭该连接
	defer redisConn.Close()

	rkey := fmt.Sprintf("%s:%X:%X", codes.RedisUserChatListPrefix, myUserID, chatID)

	// 从redis中获取所有消息
	replyBytes, err := redis.Values(redisConn.Do("HGETALL", rkey))
	if err != nil {
		return
	}

	// 有数据库 这是一个json

	info = new(api.ChatInfoDataType)

	// 将redis返回的数据包装成一个对象
	// 假如replyBytes为空会抛出error
	err = redis.ScanStruct(replyBytes, info)

	//err = json.Unmarshal(replyBytes, resp)
	if err == nil {
		// 判断是否为空
		// ChatID不能为0
		if info.ChatId != 0 {
			// 直接返回

			return
		}
	}
	// 当replyBytes为空，代表着用户与聊天对象未曾聊过天，是新的会话
	db := d.db
	// 不存在 则创建会话
	// TODO 查找用户信息 ， 后期该redis
	itemChatInfoItem := &models.ChatInfoDataType{}
	itemChatInfoItem.MyUserID = myUserID
	itemChatInfoItem.ChatID = chatID

	if chatID < 0 {
		//TODO 这是群
		// 处理群的chat info

		itemChatInfoItem.ChatTitle = fmt.Sprintf("chat title : %d", chatID)
	} else {
		// 这是用户
		uinfo := models.UserInfoViewer{}

		if dbResp := db.Model(&uinfo).Where(&models.UserInfoViewer{UserID: chatID}).Find(&uinfo); dbResp.Error != nil || dbResp.RowsAffected == 0 {
			// 查询用户失败
			d.logger.Debug("查询用户失败", zap.Error(dbResp.Error))
			err = errors.New("用户信息不存在")
			return
		}
		itemChatInfoItem.ChatName = uinfo.Nick
		if itemChatInfoItem.ChatTitle == "" {
			itemChatInfoItem.ChatName = fmt.Sprintf("p_%s", uinfo.Username)
		}
		if itemChatInfoItem.ChatTitle == "" {
			itemChatInfoItem.ChatName = fmt.Sprintf("pid_%d", uinfo.UserID)
		}

	}

	// 插入数据库
	if dbResp := db.Clauses(&clause.OnConflict{
		DoNothing: false,
	}).Create(itemChatInfoItem); dbResp.Error != nil {
		// 创建会话失败
		err = errors.New("创建会话失败")
		return
	}

	// 创建成功 存redis

	info.MyUserId = itemChatInfoItem.MyUserID
	info.ChatId = itemChatInfoItem.ChatID
	info.ChatTitle = itemChatInfoItem.ChatTitle
	info.LastUpdateTime = itemChatInfoItem.LastUpdateTime
	info.LastMsgId = itemChatInfoItem.LastMsgID

	// 同样的要重新更新

	d.AddValue(rkey, info)
	//_, err = redisConn.Do("HMSET", redis.Args{rkey}.AddFlat(info)...)
	//if err != nil {
	//	// 保存redis 失败 ??? 处理
	//	d.logger.Debug("保存会话到redis 失败")
	//}
	//_, err = redisConn.Do("EXPIRE", rkey, 3600*24*7)

	//info = itemChatInfoItem
	err = nil
	return
}

func (d *Dao) GetCacheKey(key string, update ...bool) (interface{}, error) {
	//TODO implement me
	//panic("implement me")
	d.rw.RLock()
	defer d.rw.RUnlock()

	value, isok := d.cacheGlobalMap[key]
	if isok {
		if value.Value != nil {
			return value.Value, nil
		} else {
			// 顺便清理这个key
			delete(d.cacheGlobalMap, key)
			return nil, errors.New("value is empty")
		}
	} else {

		return nil, errors.New("not cache")
	}

}

func (d *Dao) AutoClearService() {
	//TODO implement me
	//panic("implement me")

	ticker := time.NewTicker(time.Minute)

	for true {
		select {
		case timeNow := <-ticker.C:
			d.DoClearKey(timeNow)
			continue
		case <-d.closeServer:
			d.logger.Warn("AutoClearService stop")
			return

		}
	}

}

func (d *Dao) GetUserInfoByID(userid int64) (info *models.UserInfoViewer, err error) {
	//TODO implement me
	//panic("implement me")
	db := d.db

	info = new(models.UserInfoViewer)
	err = db.Model(&models.UserInfoViewer{}).Where(&models.UserInfoViewer{
		UserID: userid,
	}).Find(info).Error
	return
}

func (d *Dao) DoClearKey(timeNow time.Time) {
	tools.HandlePanic(d.logger, "DaoClearService ")

	// 清理1小时前的数据

	lastTime := timeNow.Unix() - 3600

	d.rw.Lock()
	defer d.rw.Unlock()
	for key, value := range d.cacheGlobalMap {
		if value.UpdateAt < lastTime {
			delete(d.cacheGlobalMap, key)
		}
	}

}

func (d *Dao) AddValue(key string, value interface{}) {
	valueWarp := NewValue(value)
	d.rw.Lock()
	defer d.rw.Unlock()
	d.cacheGlobalMap[key] = valueWarp
}

func NewDao(logger *zap.Logger, db *gorm.DB, redisPool *redis.Pool, closeState <-chan struct{}) APIDao {
	return &Dao{
		closeServer:    closeState,
		logger:         logger,
		db:             db,
		redisPool:      redisPool,
		rw:             new(sync.RWMutex),
		cacheGlobalMap: make(map[string]*CacheValue, 1024),
	}
}

type SystemControlDao interface {
	AutoClearService()
	GetCacheKey(key string, update ...bool) (interface{}, error)
	DoClearKey(timeNow time.Time)
}

type UserDao interface {
	// GetUserInfoByID 通过ID获取用户个人信息
	GetUserInfoByID(userid int64) (info *models.UserInfoViewer, err error)
	// GetChatInfoByID 通过当前用户ID与聊天对象ID查询聊天记录
	GetChatInfoByID(myUserID int64, chatID int64) (info *api.ChatInfoDataType, err error)
}

type APIDao interface {
	UserDao
	SystemControlDao
}
