package models

import "gorm.io/gorm"

// GroupBaseInfo 群基本信息表
type GroupBaseInfo struct {
	GroupID    int64          `json:"group_id" gorm:"type:bigint;primarykey;comment:基础id"`
	CreatedAt  int64          `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt  int64          `json:"-" gorm:"autoUpdateTime:milli"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Name       string         `json:"name" gorm:"size:32"` // 限制 32 长度的群名
	CreateUser int64          `json:"create_user"`         // 创建者
}

func (g GroupBaseInfo) TableName() string {
	return "group_base_info"
}

// GroupMember 群成员信息
type GroupMember struct {
	GroupID   int64  `json:"group_id" gorm:"primarykey;comment:基础id"`
	MemberID  int64  `json:"member_id" gorm:"primarykey"` // 使用联合主键 保证 gid - uid 唯一
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `json:"-" gorm:"autoUpdateTime:milli"`
	DeleteAt  int64  `json:"-" gorm:"primarykey"` // 把删除时间并入 为了 能够实现软删除 并且唯一 删除 则改成 update DeleteAt
	UserNote  string `json:"user_note"`           // 用户再群里的备注
	//	... 其他共用参数
	UserType int    `json:"user_type"` // 用户类型 	GroupUserTypeNormal 	GroupUserTypeAdmin 	GroupUserTypeCreator
	Nick     string `json:"nick"`      //入群时的用户信息
	Inviter  int64  `json:"inviter"`   // 邀请人

}

func (g GroupMember) TableName() string {
	return "group_members"
}

type GroupNotification struct {
	//NotificationID int64  `json:"notification_id" gorm:"primarykey;comment:基础id"`
	GroupID   int64  `json:"group_id" gorm:"comment:群id"`
	Content   string `json:"content" gorm:"comment:公告内容"`
	CreatedBy int64  `json:"created_by" gorm:"comment:创建人ID"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `json:"-" gorm:"autoUpdateTime:milli"`
	DeleteAt  int64  `json:"-" `
	Nick      string `json:"nick" gorm:"comment:创建人昵称" ` //入群时的用户信息
}

func (g GroupNotification) TableName() string {
	return "group_notifications"
}
