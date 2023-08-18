package pim_server

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"pim/api"
	"pim/pkg/models"
	"time"
)

// CreateGroup 创建群聊
func (p *PimServer) CreateGroup(ctx context.Context, req *api.CreateGroupReq) (resp *api.CreateGroupResp, err error) {
	// TODO 鉴权
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 鉴权失败，return
	// 鉴权成功
	db := p.svr.db
	logger := p.svr.logger
	// 插入群信息
	groupBaseInfo := models.GroupBaseInfo{
		Name: req.Name,
	}
	// 创建成功后，主键的值将会被插入groupBaseInfo中
	groupCreatedTime := time.Now().Unix()
	err = db.Create(&groupBaseInfo).Error
	if err != nil {
		logger.Error("群信息插入失败", zap.Int64("stream_id", req.StreamID))
		return
	}
	// 插入群主信息
	var masterInfo models.GroupMember
	err = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&masterInfo).Error
	if err != nil {
		logger.Error("群主ID有误", zap.Int64("user_id", p.clients[req.StreamID].UserID), zap.Int64("stream_id", req.StreamID))
		return
	}

	groupMasterInfo := &models.GroupMember{
		GroupID:  groupBaseInfo.GroupID,
		MemberID: masterInfo.MemberID,
		Nick:     masterInfo.Nick,
		//UserType: codes.GroupUserTypeMaster,
	}
	db.Create(groupMasterInfo)
	p.pushCacheToGroups(groupBaseInfo.GroupID, masterInfo.MemberID)
	// 插入群成员信息
	var membersInfo []*models.GroupMember
	_ = db.Where("user_id in ?", req.Members).Find(&membersInfo).Error

	if len(membersInfo) != 0 {
		var groupMemberList []*models.GroupMember
		updateGroupNewMemberDataTypeModel := api.UpdateGroupNewMemberDataType{
			UpdatedAt: groupCreatedTime,
			//InvitedBy: groupMasterInfo.Nick,
		}

		for _, m := range membersInfo {
			temp := &models.GroupMember{
				GroupID:  groupBaseInfo.GroupID,
				MemberID: m.MemberID,
				Nick:     m.Nick,
				Inviter:  tokenInfo.GetUserID(),
				//UserType: codes.GroupUserTypeNormal,
			}
			groupMemberList = append(groupMemberList, temp)
			p.pushCacheToGroups(groupBaseInfo.GroupID, temp.MemberID)

			//updateGroupNewMemberDataTypeModel.MemberNick = temp.Nick
			updateGroupNewMemberBody, _ := anypb.New(&updateGroupNewMemberDataTypeModel)
			updateGroupNewMemberPushedData := &api.UpdateEventDataType{
				Type: api.UpdateEventDataType_UpdateGroupNewMember,
				Body: updateGroupNewMemberBody,
			}
			p.UserStreamClientMap.PushUserEvent(temp.MemberID, updateGroupNewMemberPushedData)
			p.UserStreamClientMap.PushUserEvent(groupMasterInfo.MemberID, updateGroupNewMemberPushedData)
		}
		db.Create(groupMemberList)

	}
	resp = new(api.CreateGroupResp)
	// TODO 向群成员推送新聊天事件
	chatInfo := &api.ChatInfoDataType{
		ChatName:       groupBaseInfo.Name,
		ChatTitle:      groupBaseInfo.Name,
		ChatId:         groupBaseInfo.GroupID,
		LastUpdateTime: groupCreatedTime,
	}

	newChatInfoDataType := api.NewChatInfoDataType{
		ChatInfo: chatInfo,
	}

	newChatInfoBody, _ := anypb.New(&newChatInfoDataType)
	newChatInfoPushedData := &api.UpdateEventDataType{
		Type: api.UpdateEventDataType_NewChatInfo,
		Body: newChatInfoBody,
	}

	// NOTE 这里有bug , member 入参的时候无法确定合法性 ,
	//for _, m := range req.Members {
	//	p.UserStreamClientMap.PushUserEvent(m, newChatInfoPushedData)
	//}
	//
	// 以处理后的成员信息 为准 , 因为有的成员被拉黑之类的 , req.Member 甚至有问题的id
	// 有的任务屏蔽了信息 所以以处理后的目标成员为准
	for _, member := range membersInfo {
		p.UserStreamClientMap.PushUserEvent(member.MemberID, newChatInfoPushedData)
	}

	p.UserStreamClientMap.PushUserEvent(groupMasterInfo.MemberID, newChatInfoPushedData)
	// TODO 向群成员推送"欢迎加入"事件
	//body := api
	resp = new(api.CreateGroupResp)
	return
}

// GroupJoinByID 通过群ID加入群聊
func (p *PimServer) GroupJoinByID(ctx context.Context, req *api.GroupJoinByIDReq) (resp *api.BaseOk, err error) {
	// TODO 鉴权
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 鉴权未通过，return
	// 鉴权成功
	db := p.svr.db
	logger := p.svr.logger
	// 查找群
	var group models.GroupBaseInfo
	err = db.Where("group_id = ?", req.GroupID).Take(&group).Error
	// 失败，return
	if err != nil {
		logger.Error("查无此群", zap.Int64("group_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
		return
	}
	// NOTE 等自己加入后自己也是成员 所以可以直接推送
	// 查找所有群成员
	var oldGroupMembers []*models.GroupMember
	_ = db.Where("group_id = ?", req.GroupID).Find(&oldGroupMembers).Error
	// 添加群成员信息
	var this models.UserInfoViewer
	// NOTE 这里直接可以从token 获取到用户信息
	//err = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&this).Error
	// 正确的写法
	err = db.Where("user_id = ?", tokenInfo.GetUserID()).Take(&this).Error
	if err != nil {
		logger.Error("查询用户失败", zap.Int64("user_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
		return
	}
	thisGroupMember := models.GroupMember{
		GroupID:  group.GroupID,
		MemberID: this.UserID,
		//UserType: codes.GroupUserTypeNormal,
		Nick: this.Nick,
	}
	db.Create(&thisGroupMember)

	// TODO 向所有成员推送"新人入群"通知
	if len(oldGroupMembers) != 0 {

	}
	// TODO 推送当前用户新聊天事件
	resp = new(api.BaseOk)
	return
}

// GroupInviteMembers 群成员邀请新成员
func (p *PimServer) GroupInviteMembers(ctx context.Context, req *api.GroupInviteMembersReq) (resp *api.BaseOk, err error) {
	// TODO 鉴权
	// 鉴权失败
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 鉴权成功
	// 基本类似GroupJoinByID
	db := p.svr.db
	logger := p.svr.logger
	// 查找群
	var group models.GroupBaseInfo
	err = db.Where("group_id = ?", req.GroupID).Take(&group).Error
	// 失败，return
	if err != nil {
		logger.Error("查无此群", zap.Int64("group_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
		return
	}
	// 添加群成员信息
	//var thisMembers []*models.UserInfoViewer
	//err = db.Where("user_id in ?", req.Members).Find(&thisMembers).Error
	//if err != nil {
	//	logger.Error("添加群成员信息失败", zap.Int64("group_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
	//	return
	//}
	//var groupMembers []*models.GroupMember
	//if len(thisMembers) != 0 {
	//	for _, am := range thisMembers {
	//		temp := models.GroupMember{
	//			GroupID:  group.GroupID,
	//			MemberID: am.UserID,
	//			Nick:     am.Nick,
	//			Inviter:  tokenInfo.GetUserID(), // 增加一个 邀请人 业务会更完整 , 还有进群方式类型等...
	//			UserType: int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeNormal),
	//			//UserType: codes.GroupUserTypeNormal,
	//		}
	//		groupMembers = append(groupMembers, &temp)
	//
	//		// 同时添加一条入群消息
	//
	//
	//	}
	//}
	//_ = db.Create(&groupMembers)
	//// TODO 向新用户们推送新聊天事件
	//// 查找所有群成员
	//var allGroupMembers []*models.GroupMember
	//_ = db.Where("group_id = ?", req.GroupID).Find(&allGroupMembers).Error
	//// TODO 向所有成员推送"新人入群"通知
	//if len(allGroupMembers) != 0 {
	//	//for _, m := range groupMembers{
	//	//}
	//}
	//resp = new(api.BaseOk)

	// 判断群里是否有这个邀请者

	var inviterInfoByGroup models.GroupMember

	findMemberErr := db.Model(&inviterInfoByGroup).Where(&models.GroupMember{
		GroupID:  group.GroupID,
		MemberID: tokenInfo.GetUserID(),
	}).Find(&inviterInfoByGroup).Error

	if findMemberErr != nil || inviterInfoByGroup.MemberID == 0 || inviterInfoByGroup.MemberID != tokenInfo.GetUserID() {
		logger.Info("校验群成员信息失败", zap.Error(findMemberErr))
		err = errors.New("邀请入群失败,没有权限操作")
		return
	}

	// 验证 群成员是否有效
	var checkMemberUserInfo []*models.UserInfoViewer

	findErr := db.Model(&models.UserInfoViewer{}).Where("user_id in ?", req.Members).Find(&checkMemberUserInfo).Error

	if findErr != nil || len(checkMemberUserInfo) == 0 {
		logger.Error("find checkMemberUserInfo fail", zap.Error(findErr), zap.Int("count", len(checkMemberUserInfo)))
		err = errors.New("邀请用户错误,稍后重试")
		return
	}

	// 这里过滤后的群成员和群信息

	//  读取现有的成员信息

	var currentGroupMembers []*models.GroupMember

	findMemberErr = db.Model(&models.GroupMember{}).Where(&models.GroupMember{
		GroupID: group.GroupID,
	}).Find(&currentGroupMembers).Error
	// 正常邀请的群不可能没有人
	if findMemberErr != nil || len(currentGroupMembers) == 0 {
		logger.Error("群成员信息错误", zap.Error(findMemberErr))
		err = errors.New("群信息异常")
		// todo  后期需要清理服务
		return
	}

	// 这里就应该是没问题的了 , 具体操作业务要开个任务处理复杂的流程

	go runGroupMemberInviteProc(p, tokenInfo, group, currentGroupMembers, checkMemberUserInfo)

	resp = new(api.BaseOk)

	return
}

// 这是耗时的操作
func runGroupMemberInviteProc(p *PimServer, tokenInfo TokenInfo, group models.GroupBaseInfo, currentGroupMembers []*models.GroupMember, checkMemberUserInfo []*models.UserInfoViewer) {

	// 批量推送消息 循环推送
	// 参照微信 , 先入的能够收到后面的消息

	// 协议个推送的方法
	pushMessageFunc := func(groupInfo *models.GroupMember, msg *api.Message) {
		//
		//生成群信息

		newGroupEvent := &api.UpdateGroupNewMemberDataType{
			UpdatedAt: groupInfo.UpdatedAt,
			//MemberNick: groupInfo.Nick,
			InvitedBy: tokenInfo.GetUserID(),
			MemberID:  groupInfo.MemberID,
			MessageID: msg.ID,
		}

		groupNewEventBody, _ := anypb.New(newGroupEvent)

		groupNewEvent := &api.UpdateEventDataType{
			Type: api.UpdateEventDataType_UpdateGroupNewMember,
			Body: groupNewEventBody,
		}

		for _, member := range currentGroupMembers {
			// 找到成员的 客户端

			findUserClient, isok := p.UserStreamClientMap[member.MemberID]

			if !isok {
				// 用户不在线
				continue
			}

			newMessageEventbBody, _ := anypb.New(msg)

			newMessageEvent := &api.UpdateEventDataType{
				Type: api.UpdateEventDataType_NewMessage,
				Body: newMessageEventbBody,
			}

			// 推送消息
			findUserClient.PushUserEvent(newMessageEvent)
			// 先推消息可以减少一次消息的检索 因为成员信息里有个msg id
			// 先落地 那么本地就有这条数据了 可以先缓存

			// 推送 新成员数据
			findUserClient.PushUserEvent(groupNewEvent)

		}

	}
	_ = pushMessageFunc

	// 遍历邀请的用户信息
	db := p.svr.db
	logger := p.svr.logger
	// 用于生成一条需要发送的消息 其中一条是存数据库的
	genAddMemberMessageToDB := func(userInfo *models.UserInfoViewer) *api.Message {
		genMsgID := p.GenMsgID()

		newMessage := &models.SingleMessage{
			MsgID:     genMsgID.Int64(),
			CreatedAt: genMsgID.Time(),
			ChatID:    GetChatIDByBaseGroupID(group.GroupID), // 注意 群的规则是 负号
			Sender:    tokenInfo.GetUserID(),
			MsgType:   int(api.MessageTypeEnum_MessageTypeNewMember),
			MsgStatus: int(api.MessageStatusEnum_MessageStatusSend),
		}

		pushMessage := &api.Message{}
		pushMessage.Sender = newMessage.Sender
		pushMessage.ChatID = newMessage.ChatID
		pushMessage.ID = newMessage.MsgID
		pushMessage.Type = api.MessageTypeEnum_MessageTypeNewMember
		pushMessage.Status = api.MessageStatusEnum_MessageStatusSend

		p.svr.saveMessageChan <- newMessage
		//p.svr.sendMessageChan <- pushMessage

		return pushMessage
	}

	for _, userViewer := range checkMemberUserInfo {
		// 	按照顺序生成 消息

		// 生成成员信息

		newMemberInfo := &models.GroupMember{
			GroupID:  group.GroupID,
			MemberID: userViewer.UserID,
			Nick:     userViewer.Nick,
			Inviter:  tokenInfo.GetUserID(), // 增加一个 邀请人 业务会更完整 , 还有进群方式类型等...
			UserType: int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeNormal),
			//UserType: codes.GroupUserTypeNormal,
		}

		// 添加到群表

		addGroupErr := db.Create(newMemberInfo).Error
		if addGroupErr != nil {
			// 添加失败
			logger.Error("添加成员信息失败", zap.Error(addGroupErr))
			continue
		}

		// 添加成功推送一个 入群事件

		pushMessage := genAddMemberMessageToDB(userViewer)

		// 将消息推送给群里的所有人
		pushMessageFunc(newMemberInfo, pushMessage)
		// 推玩就要吧当前用户加到列表
		// 下一轮则也会推送这个用户的数据
		currentGroupMembers = append(currentGroupMembers, newMemberInfo)

	}
}
func GetChatIDByBaseGroupID(groupID int64) int64 {
	return 0 - groupID
}

func (p *PimServer) GroupEditNotification(ctx context.Context, req *api.GroupEditNotificationReq) (resp *api.BaseOk, err error) {
	//todo
	// 鉴权
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	db := p.svr.db
	logger := p.svr.logger
	// 用户是否有权限编辑通知
	// 否，return
	var thisUserInfoViewer models.UserInfoViewer
	_ = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&thisUserInfoViewer).Error
	//if thisUserInfoViewer.UserType == codes.GroupUserTypeNormal
	if thisUserInfoViewer.UserType == 0 {
		logger.Info("用户无权限增加群通知", zap.Int64("user_id", thisUserInfoViewer.UserID))
		return
	}
	// 向所有用户推送通知
	var allGroupMembers []*models.GroupMember
	_ = db.Where("group_id = ?", req.GroupID).Find(&allGroupMembers).Error
	if len(allGroupMembers) != 0 {
		//for _, m := range allGroupMembers {
		//
		//}
	}
	return
}

func (p *PimServer) GroupRemoveMembers(ctx context.Context, req *api.GroupRemoveMembersReq) (resp *api.BaseOk, err error) {
	// 鉴权
	// 鉴权失败
	tokenInfo, err := p.CheckAuthByStream(req)
	if err != nil {
		return
	}
	// 用户信息的使用
	_ = tokenInfo
	// 鉴权成功
	db := p.svr.db
	logger := p.svr.logger
	// 是否是管理员或群主
	// 否，return
	var thisUserInfoViewer models.UserInfoViewer
	_ = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&thisUserInfoViewer).Error
	//if thisUserInfoViewer.UserType == codes.GroupUserTypeNormal
	if thisUserInfoViewer.UserType == 0 {
		logger.Info("用户无权限删除用户", zap.Int64("user_id", thisUserInfoViewer.UserID))
		return
	}
	// 删除群成员信息
	// TODO 处理复杂化
	// 1. 数组没有定义长度的情况下 deletedGroupMembers[k]
	// 因为 deletedGroupMembers = nil ,
	// 所以 deletedGroupMembers[1] 是空指针 , 直接程序 panic
	// 删除 完全可以 使用 where  user_id in [1,2,3]  这样的sql 语句解决
	// gorm 直接 db.Where("user_id in ? " , []int{} ).Delete(...)
	//var deletedGroupMembers []models.GroupMember
	//for k, m := range req.Members {
	//  这里会崩
	//	deletedGroupMembers[k] = models.GroupMember{
	//		MemberID: m,
	//	}
	//}
	//_ = db.Delete(&deletedGroupMembers).Error
	// 删除群成员对应缓存
	// 向被删除的成员推送"已被移出群聊信息"
	resp = new(api.BaseOk)
	return
}

type SimpleGroupMembersStruct []int64

// groupID ->
type GroupCache map[int64]SimpleGroupMembersStruct

// 缓存群成员，不用每次都找
// 可是如果每个用户都被缓存在内存可能空间不够
func (p *PimServer) pushCacheToGroups(groupID int64, values ...int64) {
	p.groups[groupID] = append(p.groups[groupID], values...)
}

func (p *PimServer) convertStreamIDToUserID(streamID int64) int64 {
	//TODO  这里有问题 ,如果streamID 找不到 数据
	// p.clients[streamID] 为 nil 时
	// 没有 UserID 成员
	//return p.clients[streamID].UserID
	// 正确写法是

	findClient, isok := p.clients[streamID]
	if isok {
		return findClient.UserID
	} else {
		return 0
	}

}
