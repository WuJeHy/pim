package pim_server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
	"pim/api"
	"pim/pkg/codes"
	"pim/pkg/models"
)

func (p *PimServer) CrateChat(ctx context.Context, req *api.UserIDReq) (resp *api.ChatInfoDataType, err error) {

	tokenInfo, err := p.CheckAuthByStream(req)

	if err != nil {
		return
	}

	logger := p.svr.logger
	if req.UserID <= 0 {
		err = errors.New("userid fail")
		return
	}

	redisConn := p.svr.redisPool.Get()
	defer redisConn.Close()

	rkey := fmt.Sprintf("%s:%X:%X", codes.RedisUserChatListPrefix, tokenInfo.GetUserID(), req.UserID)

	replyBytes, err := redis.Values(redisConn.Do("HGETALL", rkey))
	if err != nil {
		return
	}

	// 有数据库 这是一个json

	resp = new(api.ChatInfoDataType)

	err = redis.ScanStruct(replyBytes, resp)

	//err = json.Unmarshal(replyBytes, resp)
	if err == nil {
		// 判断是否为空
		if resp.ChatId != 0 {
			// 直接返回

			return
		}
	}

	// 不存在 则创建会话
	// TODO 查找用户信息 ， 后期该redis

	uinfo := models.UserInfoViewer{}

	db := p.svr.db

	if dbResp := db.Model(&uinfo).Where(&models.UserInfoViewer{UserID: req.UserID}).Find(&uinfo); dbResp.Error != nil || dbResp.RowsAffected == 0 {
		// 查询用户失败
		logger.Debug("查询用户失败", zap.Error(dbResp.Error))
		err = errors.New("用户信息不存在")
		return
	}

	itemChatInfoItem := &models.ChatInfoDataType{}
	itemChatInfoItem.MyUserID = tokenInfo.GetUserID()
	itemChatInfoItem.ChatID = req.UserID
	itemChatInfoItem.ChatName = uinfo.Nick
	if itemChatInfoItem.ChatTitle == "" {
		itemChatInfoItem.ChatName = fmt.Sprintf("p_%s", uinfo.Username)
	}
	if itemChatInfoItem.ChatTitle == "" {
		itemChatInfoItem.ChatName = fmt.Sprintf("pid_%d", uinfo.UserID)
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
	resp.MyUserId = itemChatInfoItem.MyUserID
	resp.ChatId = itemChatInfoItem.ChatID
	resp.ChatName = itemChatInfoItem.ChatName
	resp.LastUpdateTime = itemChatInfoItem.LastUpdateTime
	resp.LastMsgId = itemChatInfoItem.LastMsgID
	_, err = redisConn.Do("HMSET", redis.Args{rkey}.AddFlat(resp)...)
	if err != nil {
		// 保存redis 失败 ??? 处理
		logger.Debug("保存会话到redis 失败")
	}
	_, err = redisConn.Do("EXPIRE", rkey, 3600*24*7)

	err = nil
	return
}

func (p *PimServer) SendMessage(ctx context.Context, req *api.SendMessageReq) (resp *api.SendMessageResp, err error) {
	// 发送消息

	tokenInfo, err := p.CheckAuthByStream(req)

	if err != nil {
		return
	}

	if req.ChatID > 0 {
		//这是私聊的消息
	}

	msgID := p.GenMsgID()
	//
	createAt := msgID.Time()

	paramJson, _ := json.Marshal(req.Params)
	atUserJson, _ := json.Marshal(req.AtUser)
	newMessageID := msgID.Int64()
	// 产生消息
	saveMsg := &models.SingleMessage{
		MsgID:            newMessageID,
		CreatedAt:        createAt,
		UpdatedAt:        createAt,
		Sender:           tokenInfo.GetUserID(),
		ChatID:           req.ChatID,
		MsgType:          int(req.GetType()),
		MsgStatus:        int(api.MessageStatusEnum_MessageStatusSend),
		Text:             req.MessageText,
		Params:           paramJson,
		AtUser:           atUserJson,
		ReplyToMessageID: req.ReplyToMessageID,
		ReplyInChatID:    req.ReplyInChatID,
		//Body:             []byte(req.MessageText),
		//Attach:           req.Attach,
	}

	sendMsg := &api.Message{
		ID:               newMessageID,
		CreatedAt:        createAt,
		UpdatedAt:        createAt,
		Sender:           tokenInfo.GetUserID(),
		ChatID:           req.ChatID,
		Type:             req.GetType(),
		MessageText:      req.GetMessageText(),
		ReplyToMessageID: req.GetReplyToMessageID(),
		ReplyInChatID:    req.GetReplyInChatID(),
		AtUser:           req.GetAtUser(),
		Status:           api.MessageStatusEnum_MessageStatusSend,
	}

	// 推送给我和对方

	p.svr.saveMessageChan <- saveMsg
	p.svr.sendMessageChan <- sendMsg

	//sendMsg := &models.SingleMessageDataType{
	//	MsgID:            newMessageID,
	//	CreatedAt:        createAt,
	//	UpdatedAt:        createAt,
	//	Sender:           c.UserID,
	//	ChatID:           req.ChatID,
	//	MsgType:          req.MsgType,
	//	MsgStatus:        models.SenderMsgStateSend,
	//	Body:             sav,
	//	Attach:           req.Attach,
	//	Params:           req.Params,
	//	AtUser:           req.AtUser,
	//	ReplyToMessageID: req.ReplyToMessageID,
	//	ReplyInChatID:    req.ReplyInChatID,
	//}

	resp = new(api.SendMessageResp)

	resp.ID = newMessageID

	return
}
