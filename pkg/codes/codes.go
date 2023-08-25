package codes

const (
	UserMod         = 1
	MsgMod          = 2 // 消息模块
	EventMod        = 3
	ContactMod      = 4
	ConversationMod = 5
	GroupMod        = 6
	TestMod         = 99
)
const (
	GroupUserTypeNormal  = 0 // 普通用户
	GroupUserTypeAdmin   = 1 // 管理者者
	GroupUserTypeCreator = 2 // 创建者

)

const (
	UserModGetMeReq             = 1
	UserModGetMeResp            = 2
	UserModGetUserByIdReq       = 3
	UserModGetUserByIdResp      = 4
	UserModAddUserToContactReq  = 5
	UserModAddUserToContactResp = 6
)

const (
	GroupModCreateGroupReq  = 1
	GroupModCreateGroupResp = 2
	GroupModNewEntrantsReq  = 3
	GroupModNewEntrantsResp = 4
)

const (
	ConversationModGetListReq  = 1
	ConversationModGetListResp = 2
)

const (
	ContactModAdd = 1
)
const (
	TestModPing = 1
)

const (
	MsgModSenderMsgReq              = 1
	MsgModSenderMsgResp             = 2
	MsgModCreateChatReq             = 3
	MsgModCreateChatResp            = 4
	MsgModSyncChatReq               = 5
	MsgModSyncChatResp              = 6
	MsgModGetHistoryMsgReq          = 7
	MsgModGetHistoryMsgResp         = 8
	MsgModCreateUploadTaskReq       = 9
	MsgModCreateUploadTaskResp      = 10
	MsgModStopUploadTaskReq         = 11
	MsgModStopUploadTaskResp        = 12
	MsgModSendUploadPackageReq      = 13
	MsgModSendUploadPackageResp     = 14
	MsgModContinueUploadPackageReq  = 15
	MsgModContinueUploadPackageResp = 16
	MsgModPauseUploadPackageReq     = 17
	MsgModPauseUploadPackageResp    = 18
	MsgModDoneUploadPackageReq      = 19
	MsgModDoneUploadPackageResp     = 20
)

const (
	EventModConnectSuccess = 1
	EventModConnectFail    = 2
	EventModNewMessage     = 3
	EventModKickDevice     = 4
	EventModUpdateUserInfo = 5
	EventModNewChatInfo    = 6
)

const (
	SenderMessageTypeText  = 1
	SenderMessageTypeImage = 2
)

// Redis前缀
const (
	RedisUserChatListPrefix = "UCI"
	RedisGroupMemberPrefix  = "GMB"
)
