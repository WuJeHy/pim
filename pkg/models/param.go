package models

type LimitOffsetReq struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// ChatInfoDataType 聊天信息放 redis 缓存
type ChatInfoDataType struct {
	ChatInfo
	LastUpdateTime int64 `json:"last_update_time" gorm:"column:last_update_time"` // 最后更新时间
	LastMsgID      int64 `json:"last_msg_id"  gorm:"column:last_msg_id"`          // 最后一条聊天记录的ID
}

func (r ChatInfoDataType) TableName() string {
	return "chat_infos"
}

// ChatInfo 聊天信息列表
// A -> ChatID
type ChatInfo struct {
	ChatID    int64  `json:"chat_id" gorm:"primaryKey;autoIncrement:false"` // 聊天对象ID
	MyUserID  int64  `json:"-" gorm:"primaryKey;autoIncrement:false"`       // 自己的ID
	ChatName  string `json:"chat_name"`                                     // 聊天名称（假如是私聊，这个就是对方名字）
	ChatTitle string `json:"chat_title"`                                    // 聊天标题（群聊中，可以是群名，也可以是自己起的备注）
}

type ConversionItem struct {
	Count    int                 `json:"count"`
	ChatList []*ChatInfoDataType `json:"chats"`
}
