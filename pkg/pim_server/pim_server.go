package pim_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm/clause"
	"pim/api"
	"pim/pkg/models"
	"pim/pkg/tools"
	"sync"
	"time"
)

type RpcClient struct {
	//	一个链接的客户端
	UserID   int64
	Pf       int
	StreamID int64
	Level    int
	PushFunc func(event *api.UpdateEventDataType)
}

// GetUserID 实现token 的接口
func (c *RpcClient) GetUserID() int64 {
	return c.UserID
}

func (c *RpcClient) GetPf() int {
	return c.Pf
}

func (c *RpcClient) GetLevel() int {
	return c.Level
}

type StreamInfoReq interface {
	GetStreamID() int64
}

// CheckAuthByStream 通过流id 鉴权
func (p *PimServer) CheckAuthByStream(req StreamInfoReq) (token TokenInfo, err error) {
	p.rw.RLock()
	defer p.rw.RUnlock()

	client, isok := p.clients[req.GetStreamID()]
	if !isok {
		err = errors.New("没有权限")
		return
	}

	token = client

	return
}

// StreamClientType 类型别名
type StreamClientType map[int64]*RpcClient

type PimServer struct {
	svr *server
	// 这个map 是调用接口的时候快速查询用的
	clients map[int64]*RpcClient
	rw      *sync.RWMutex
	// TODO 用户映射还没加
	UserStreamClientMap map[int64]StreamClientType
}

// 初始化业务
func (c *RpcClient) initClient() {
	// 推送流
	pushConnectSuccessEvent := &api.ConnectSuccessDataType{
		StreamID: c.StreamID,
	}
	body, _ := anypb.New(pushConnectSuccessEvent)

	pushData := &api.UpdateEventDataType{
		Type: api.UpdateEventDataType_ConnectSuccess,
		Body: body,
	}

	c.PushFunc(pushData)

}

func (p *PimServer) UpdateEvent(req *api.TokenReq, eventServer api.PimServer_UpdateEventServer) error {
	//TODO implement me
	//panic("implement me")

	// todo Token 校验
	// token 解析

	// 这里的token 是jwt 的形式处理的

	streamID := int64(uuid.New().ID())

	logger := p.svr.logger.Named(fmt.Sprintf("%d", streamID))
	tokenInfo, errToken := tools.ParseToken(req.Token)
	if errToken != nil {

		logger.Debug("token 解析错误", zap.Error(errToken))

		err := errors.New("token 校验失败")
		return err
	}

	eventChannel := make(chan *api.UpdateEventDataType, 8)
	//lock := new(sync.Mutex)
	client := &RpcClient{
		StreamID: streamID,
		Pf:       int(tokenInfo.Pf),
		UserID:   tokenInfo.UserID,
		Level:    tokenInfo.Level,
		PushFunc: func(event *api.UpdateEventDataType) {
			//lock.Lock()
			//defer lock.Unlock()
			eventChannel <- event
		},
	}

	// 这里需要添加用户关系
	//需要绑定 userid -> stream_id 的关系 即可

	p.AddUserStream(client)

	logger.Info("新用户接入", zap.Int64("UID", client.UserID))

	defer p.RemoveUserStream(client)

	// 推送

	// 推送登录成功数据
	client.initClient()

	for true {
		select {
		case <-p.svr.closeServer:
			logger.Info("主服务发出退出指令")
			return nil

		case <-eventServer.Context().Done():
			logger.Info("Exit ctx", zap.Error(eventServer.Context().Err()))
			return eventServer.Context().Err()

		case pushEvent := <-eventChannel:

			err := eventServer.Send(pushEvent)

			if err != nil {
				logger.Error("send err", zap.Error(err))
				return err
			}

		}
	}

	return nil
}

func (p *PimServer) Register(ctx context.Context, req *api.RegisterReq) (resp *api.BaseOk, err error) {
	//TODO implement me
	// 这个接口没有鉴权

	db := p.svr.db
	logger := p.svr.logger
	var authInfo models.Auth

	authInfo.Password = req.Password
	authInfo.Email = req.Email
	authInfo.Username = req.Username

	errSave := db.Create(&authInfo).Error
	if errSave != nil {

		logger.Error("注册账户失败", zap.Any("params", req), zap.Error(errSave))
		err = errors.New("注册失败")
		return
	}

	resp = new(api.BaseOk)
	//// 创建用户信息
	//var userinfo models.UserInfo
	//
	//userinfo.UserID = authInfo.UserID
	//userinfo.Username = req.Username
	//
	//err = s.db.Create(&userinfo).Error
	//
	//if err != nil {
	//	tools.Resp500(ctx, err.Error())
	//	return
	//}

	return

}

func (p *PimServer) Login(ctx context.Context, req *api.LoginReq) (resp *api.LoginResp, err error) {
	//TODO implement me
	//panic("implement me")
	// 查找 用户

	db := p.svr.db

	logger := p.svr.logger

	switch req.Type {
	case api.LoginReq_LoginByUsername:
	default:
		err = errors.New("不支持的登录类型")
		return
	}
	var userinfo models.Auth

	errFind := db.
		Model(&userinfo).
		Where("(email = ? or username = ? or mobile = ?) and password = ? ",
			req.Username, req.Username, req.Username, req.Password).
		Find(&userinfo).Error

	if errFind != nil || userinfo.UserID == 0 {
		logger.Error("查询账户失败", zap.Error(errFind))
		err = errors.New("登录失败")
		return
	}

	tokenStr, _ := tools.GenToken(userinfo.UserID, int(req.Platform), userinfo.Level)

	resp = new(api.LoginResp)
	resp.Token = tokenStr
	return

}

func (p *PimServer) GetMyUserInfo(ctx context.Context, req *api.StreamReq) (resp *api.UserInfoViewerDataType, err error) {
	// 从流中提取基本信息

	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 查询我的信息

	logger := p.svr.logger
	db := p.svr.db

	var userinfo models.UserInfoViewer

	err = db.Model(&userinfo).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Find(&userinfo).Error

	if err != nil || userinfo.UserID == 0 {
		logger.Debug("get user info by id fail ", zap.Error(err))
		err = errors.New("user not found")
		return
	}
	resp = new(api.UserInfoViewerDataType)
	resp.UserID = userinfo.UserID
	resp.Username = userinfo.Username
	resp.Email = userinfo.Email
	resp.Nick = userinfo.Nick
	resp.CreatedAt = userinfo.CreatedAt
	resp.Avatar = userinfo.Avatar
	resp.UserType = api.UserTypeEnumType(userinfo.UserType)
	resp.UserStatus = api.UserStatusEnumType(userinfo.UserStatus)

	// avatar ...

	return
}

func (p *PimServer) GetUserInfoByID(ctx context.Context, req *api.GetUserInfoByIDReq) (resp *api.UserInfoViewerDataType, err error) {
	//TODO implement me
	//panic("implement me")
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 查询目标用户的

	logger := p.svr.logger
	db := p.svr.db

	var userinfo models.UserInfoViewer

	err = db.Model(&userinfo).Where(&models.UserInfoViewer{
		UserID: req.UserID,
	}).Find(&userinfo).Error

	if err != nil || userinfo.UserID == 0 {
		logger.Debug("get user info by id fail ", zap.Error(err))
		err = errors.New("user not found")
		return
	}
	resp = new(api.UserInfoViewerDataType)
	resp.UserID = userinfo.UserID
	resp.Username = userinfo.Username
	resp.Email = userinfo.Email
	resp.Nick = userinfo.Nick
	resp.CreatedAt = userinfo.CreatedAt
	resp.Avatar = userinfo.Avatar
	resp.UserType = api.UserTypeEnumType(userinfo.UserType)
	resp.UserStatus = api.UserStatusEnumType(userinfo.UserStatus)

	// avatar ...

	return
}

func (p *PimServer) AddUserToContact(ctx context.Context, req *api.AddUserToContactReq) (resp *api.BaseOk, err error) {
	//TODO implement me
	//panic("implement me")
	resp = new(api.BaseOk)

	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo

	// 数据加入数据库

	req.UserID = tokenInfo.GetUserID()
	//

	db := p.svr.db
	logger := p.svr.logger
	// 判断用户是否存在

	var findUser models.UserInfoViewer

	//findUser.UserID = cUserID

	respDB := db.Model(&findUser).Where(&models.UserInfoViewer{
		UserID: tokenInfo.GetUserID(),
	}).Find(&findUser)

	if respDB.Error != nil {
		//err = respDB.Error
		logger.Debug("查找用户失败", zap.Error(respDB.Error))
		err = errors.New("添加失败")
		return
	}

	updateMap := map[string]interface{}{}

	//if req.FirstName != "" {
	//	updateMap["first_name"] = req.FirstName
	//}
	//
	//if req.LastName != "" {
	//	updateMap["last_name"] = req.LastName
	//}

	if req.Note != "" {
		updateMap["note"] = req.Note
	}

	timeNow := time.Now()

	updateMap["updated_at"] = timeNow.UnixMilli()

	// 可以添加
	addErr := db.Clauses(&clause.OnConflict{
		//DoNothing: false,
		DoUpdates: clause.Assignments(updateMap),
	}).Create(req).Error

	if addErr != nil {
		logger.Debug("添加用户失败", zap.Error(addErr))

		err = errors.New("添加失败")
		return
	}

	return
}

func (p *PimServer) SendMessage(ctx context.Context, req *api.SendMessageReq) (resp *api.SendMessageResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) GetClientByStream(streamID int64) (client *RpcClient, isok bool) {

	p.rw.RLock()
	client, isok = p.clients[streamID]
	p.rw.RUnlock()

	return
}

func (p *PimServer) AddUserStream(client *RpcClient) {
	p.rw.Lock()
	defer p.rw.Unlock()
	p.clients[client.StreamID] = client

	logger := p.svr.logger
	// 映射用户关系
	//查看 用户是否有其他设备
	streamClients, isok := p.UserStreamClientMap[client.UserID]
	if isok {
		streamClients[client.StreamID] = client
		logger.Info("新设备登录", zap.Int64("StreamID", client.StreamID), zap.Int("用户在线设备数", len(streamClients)))
	} else {
		// 第一个登录的设备

		streamClients = make(StreamClientType)
		streamClients[client.StreamID] = client
		p.UserStreamClientMap[client.UserID] = streamClients
		logger.Info("第一个新设备登录", zap.Int64("StreamID", client.StreamID), zap.Int("用户在线设备数", len(streamClients)))

	}

}

func (p *PimServer) RemoveUserStream(client *RpcClient) {
	p.rw.Lock()
	defer p.rw.Unlock()
	//
	// 删除user

	logger := p.svr.logger
	streamClients, isok := p.UserStreamClientMap[client.UserID]

	if isok {
		delete(streamClients, client.StreamID)
	}
	delete(p.clients, client.StreamID)

	logger.Info("用户设备离线", zap.Int64("StreamID", client.StreamID), zap.Int("用户在线设备数", len(streamClients)))
	if len(streamClients) == 0 {
		delete(p.UserStreamClientMap, client.UserID)
	}
}
