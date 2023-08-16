package pim_server

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"pim/api"
	"pim/pkg/tools"
)

// StreamClientType 类型别名
type StreamClientType map[int64]*RpcClient

// PushUserEvent 封装直接向某个用户推送
func (s StreamClientType) PushUserEvent(event *api.UpdateEventDataType) {
	for _, client := range s {
		client.PushFunc(event)
	}
}

// PushUserEventByPf 封装直接向某个用户的某个平台推送
func (s StreamClientType) PushUserEventByPf(pf int, event *api.UpdateEventDataType) {
	for _, client := range s {
		if client.Pf == pf {
			client.PushFunc(event)
		}
	}
}

type UserStreamClientMapType map[int64]StreamClientType

// PushUserEvent 直接推送给指定的用户
func (u UserStreamClientMapType) PushUserEvent(userID int64, event *api.UpdateEventDataType) {

	streamClient, isok := u[userID]
	if isok {
		streamClient.PushUserEvent(event)
	}

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

func (p *PimServer) GetClientByStream(streamID int64) (client *RpcClient, isok bool) {

	p.rw.RLock()
	client, isok = p.clients[streamID]
	p.rw.RUnlock()

	return
}
