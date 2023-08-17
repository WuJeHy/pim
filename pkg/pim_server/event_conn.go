package pim_server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"gorm.io/gorm"
	"pim/api"
	"pim/pkg/models"
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
		svr: p.svr,
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

	// 同步会话

	c.doPushConversions(c.svr.db, c.UserID)
	// 同步事件

}

func (c *RpcClient) doPushConversions(db *gorm.DB, userID int64) {
	// 读取会话列表

	//var chatList = []*models.ChatInfoDataType{}
	logger := c.svr.logger
	var chatList = make(map[int64]*api.ChatInfoDataType)
	//rest := db.Model(&models.ChatInfoDataType{})

	rest := db.Raw("select  chat.chat_id as chat_id  , chat.my_user_id as my_user_id,chat.chat_name as chat_name ,chat.chat_title as chat_title,"+
		"tmp.msg_id as last_msg_id,tmp.created_at as last_update_time "+
		" from chat_infos chat LEFT JOIN "+
		"(select all_msg.* from single_messages all_msg where all_msg.msg_id in (select MAX(msg.msg_id) "+
		"from single_messages msg where  msg.sender = ? or msg.chat_id =?  GROUP BY msg.sender ,msg.chat_id ) ORDER BY all_msg.created_at "+
		") tmp on (tmp.chat_id = chat.chat_id and tmp.sender = chat.my_user_id) or (tmp.chat_id = chat.my_user_id and tmp.sender = chat.chat_id) WHERE my_user_id = ?", c.UserID, c.UserID, c.UserID)

	//rest.Rows()
	rows, err := rest.Rows()

	if err != nil {
		return
	}

	for rows.Next() {
		item := &api.ChatInfoDataType{}
		var lastMsgID interface{}
		var lastUpdateTime interface{}
		err = rows.Scan(&item.ChatId, &item.MyUserId, &item.ChatName, &item.ChatTitle, &lastMsgID, &lastUpdateTime)
		if err != nil {
			continue
		}

		if lastMsgID != nil {
			item.LastMsgId = lastMsgID.(int64)
		}
		if lastUpdateTime != nil {
			item.LastUpdateTime = lastUpdateTime.(int64)
		}
		//判断是否存在会话
		oldChat, isok := chatList[item.ChatId]
		if !isok {
			chatList[item.ChatId] = item
			continue
		}
		// 判断那个消息最新
		if oldChat.LastMsgId < item.LastMsgId {
			// 新消息更新 则使用新的代替
			chatList[item.ChatId] = item
		}

		// 否则不处理
		//chatList = append(chatList, item)
	}

	logger.Debug("chat list", zap.Any("chat ", chatList))
	// 推送会话列表

	var lastMsgItem []int64

	// 查询
	for _, chatInfo := range chatList {
		//jsonBuffer , _ := json.Marshal(chatInfo)

		chatBody, _ := anypb.New(chatInfo)

		pushEventData := &api.UpdateEventDataType{
			Type: api.UpdateEventDataType_NewChatInfo,
			Body: chatBody,
		}

		c.PushFunc(pushEventData)

		//_ = c.PushEventToClient(codes.EventMod, codes.EventModNewChatInfo, chatInfo)
		//同时推送最新的消息到客户端
		//if chatInfo.LastMsgID > 0 {
		//	lastMsgItem = append(lastMsgItem, chatInfo.LastMsgID)
		//}
	}

	// 查询需要推送的消息到客户端

	var willPushLastMsgItem []*models.SingleMessage

	err = db.Model(&models.SingleMessage{}).Where("msg_id in ? ", lastMsgItem).Find(&willPushLastMsgItem).Error
	if err != nil {
		return
	}

	for _, message := range willPushLastMsgItem {

		//类型转化

		pushMsgObj := &api.Message{
			ID:               message.MsgID,
			ChatID:           message.ChatID,
			ReplyToMessageID: message.ReplyToMessageID,
			ReplyInChatID:    message.ReplyInChatID,
			MessageText:      message.Text,
			//ImageContent:
			//Params:
			Type: api.MessageTypeEnum(message.MsgType),
			//AtUser:
			CreatedAt: message.CreatedAt,
			UpdatedAt: message.UpdatedAt,
			Sender:    message.Sender,
			Status:    api.MessageStatusEnum(message.MsgStatus),
		}

		_ = json.Unmarshal(message.AtUser, &pushMsgObj.AtUser)
		_ = json.Unmarshal(message.Params, &pushMsgObj.Params)
		//message.AtUser.Scan(message.AtUser)
		msgBody, _ := anypb.New(pushMsgObj)
		pushEventData := &api.UpdateEventDataType{
			Type: api.UpdateEventDataType_NewMessage,
			Body: msgBody,
		}

		c.PushFunc(pushEventData)
		//pushBody := DBMessageToClientMessage(c.UserID, message)
		//
		//_ = c.PushEventToClient(codes.EventMod, codes.EventModNewMessage, pushBody)
	}

}
func (p *PimServer) GetClientByStream(streamID int64) (client *RpcClient, isok bool) {

	p.rw.RLock()
	client, isok = p.clients[streamID]
	p.rw.RUnlock()

	return
}
