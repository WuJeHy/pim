package models

import "gorm.io/gorm"
import "gorm.io/datatypes"

type SenderMsgReq struct {
	// chat 规则
	// 0x01020304 1234  全局用户id
	// 0x01020304 0000  公网用户id
	// 0x00000000 1234  子网用户id
	// -0x0112345678 0000 公网普通群
	// -0x0100000000 1234 私网普通群
	// -0x1012345678 0000 公网超级群
	// -0x1000000000 1234 私网超级群
	ChatID           int64             `json:"chat_id"` // 普通用户正数 ， 群负数 超级群 -100开头
	MsgType          int               `json:"msg_type"`
	Attach           []byte            `json:"attach"`
	Body             []byte            `json:"body"`
	AtUser           []int64           `json:"at_user"`
	ReplyToMessageID int64             `json:"reply_to_message_id"`
	ReplyInChatID    int64             `json:"reply_in_chat_id"`
	Params           map[string]string `json:"params"`
}

type SenderMsgResp struct {
	MsgID int64 `json:"msg_id"`
}

const (
	SenderMsgStateSend    = 1
	SenderMsgStateAck     = 2
	SenderMsgStateSuccess = 3
	SenderMsgStateFail    = 4
)

type SingleMessage struct {
	MsgID            int64          `form:"msg_id" json:"msg_id" gorm:"type:bigint;primarykey;comment:基础id"` // 基础id 10 位数 1000000000 开始
	CreatedAt        int64          `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt        int64          `json:"-" gorm:"autoUpdateTime:milli"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	Sender           int64          `json:"sender"`
	ChatID           int64          `json:"chat_id"`
	MsgType          int            `json:"msg_type"`   // 1 -- 文本消息 2 -- 图片消息
	MsgStatus        int            `json:"msg_status"` // 消息的状态 1 -- 发送成功 , 2 -- 接收成功 , 3 -- 确认接收
	Body             []byte         `json:"body"`
	Attach           []byte         `json:"attach"`
	Params           datatypes.JSON `json:"params"`
	AtUser           datatypes.JSON `json:"at_user"`
	ReplyToMessageID int64          `json:"reply_to_message_id"`
	ReplyInChatID    int64          `json:"reply_in_chat_id"`
}

type SingleMessageDataType struct {
	MsgID            int64             `form:"msg_id" json:"msg_id" gorm:"type:bigint;primarykey;comment:基础id"` // 基础id 10 位数 1000000000 开始
	CreatedAt        int64             `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt        int64             `json:"-" gorm:"autoUpdateTime:milli"`
	Sender           int64             `json:"sender"`
	ChatID           int64             `json:"chat_id"`
	MsgType          int               `json:"msg_type"`
	MsgStatus        int               `json:"msg_status"` // 消息的状态 1 -- 发送成功 , 2 -- 接收成功 , 3 -- 确认接收
	Body             []byte            `json:"body"`
	Attach           []byte            `json:"attach"`
	Params           map[string]string `json:"params"`
	AtUser           []int64           `json:"at_user"`
	ReplyToMessageID int64             `json:"reply_to_message_id"`
	ReplyInChatID    int64             `json:"reply_in_chat_id"`
}

type CreateChatReq struct {
	UserID int64 `json:"user_id"`
}
