package models

import "gorm.io/gorm"

type UserInfoViewer struct {
	UserID     int64  `json:"user_id" gorm:"type:bigint;primarykey;comment:基础id"`
	CreatedAt  int64  `json:"created_at" gorm:"autoCreateTime:milli"`
	Username   string `json:"username"`
	Nick       string `json:"nick"`
	Email      string `json:"email" gorm:"unique"`
	UserStatus int    `json:"user_status"`
	UserType   int    `json:"user_type"`
	Avatar     []byte `json:"avatar"`
}

func (receiver UserInfoViewer) TableName() string {
	return "auths"
}

type Auth struct {
	UserID     int64          `json:"user_id" gorm:"type:bigint;primarykey;comment:基础id"`
	CreatedAt  int64          `json:"created_at" gorm:"autoCreateTime:milli"`
	UpdatedAt  int64          `json:"-" gorm:"autoUpdateTime:milli"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	Password   string         `json:"password" `
	Email      string         `json:"email" gorm:"unique"`
	Mobile     string         `json:"mobile"`
	Level      int            `json:"level"` // 1 -- 管理员 2 -- 管理层人员 3 -- 企业员工 4 -- 普通员工 5 -- 游客
	Username   string         `json:"username" gorm:"unique"`
	Nick       string         `json:"nick"`
	UserStatus int            `json:"user_status"`
	UserType   int            `json:"user_type"`
	Avatar     []byte         `json:"avatar"`
}

type GetUserByIdReq struct {
	UserID int64 `json:"user_id"`
}
