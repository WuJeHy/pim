package pim_server

import (
	"context"
	"errors"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"pim/api"
	"pim/pkg/models"
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
	//groupCreatedTime := time.Now().Unix()
	err = db.Create(&groupBaseInfo).Error
	if err != nil {
		logger.Error("群信息插入失败", zap.Int64("stream_id", req.StreamID))
		return
	}
	// 插入群主信息
	var masterInfo models.GroupMember
	//err = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&masterInfo).Error
	err = db.Where("user_id = ?", tokenInfo.GetUserID()).Take(&masterInfo).Error
	if err != nil {
		//logger.Error("群主ID有误", zap.Int64("user_id", p.clients[req.StreamID].UserID), zap.Int64("stream_id", req.StreamID))
		logger.Error("群主ID有误", zap.Int64("user_id", tokenInfo.GetUserID()), zap.Int64("stream_id", req.StreamID))
		return
	}

	groupMasterInfo := &models.GroupMember{
		GroupID:  groupBaseInfo.GroupID,
		MemberID: masterInfo.MemberID,
		Nick:     masterInfo.Nick,
		UserType: int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeCreator),
	}
	// NOTE  错误处理 凡是修改都要处理
	addErr := db.Create(groupMasterInfo).Error

	if addErr != nil {
		logger.Error("创建群组失败", zap.Error(addErr))
		return
	}

	//p.pushCacheToGroups(groupBaseInfo.GroupID, masterInfo.MemberID)
	// 插入群成员信息
	// NOTE 刚创建群是没有成员 , 所以这一步是多余的
	//var membersInfo []*models.GroupMember
	// 这一段查询的是群成员列表
	//_ = db.Where("user_id in ?", req.Members).Find(&membersInfo).Error

	var currentMembers []*models.GroupMember

	// 一开始只有群主
	currentMembers = append(currentMembers, groupMasterInfo)

	var memberUserInfos []*models.UserInfoViewer

	findErr := db.Model(&models.UserInfoViewer{}).Where("users in ? ", req.Members).Find(&memberUserInfos).Error
	if findErr != nil || len(memberUserInfos) == 0 {
		// 可以直接结束了

		// 给成员推送消息

		// todo

		//pushGroupMemberMessageFunc()

		return &api.CreateGroupResp{
			GroupID: groupBaseInfo.GroupID,
		}, nil
	}
	// TODO 向群成员推送"欢迎加入"事件
	//body := api

	//有成员时的逻辑

	// 需要耗时操作

	// 业务逻辑同邀请相同 ,邀请人群主而已

	go runGroupMemberInviteProc(p, tokenInfo, groupBaseInfo, currentMembers, memberUserInfos)

	resp = new(api.CreateGroupResp)
	resp.GroupID = groupBaseInfo.GroupID
	// NOTE 这里也没有返回 群id 客户端没办法处理这个接口
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
	//// NOTE 等自己加入后自己也是成员 所以可以直接推送
	//// 查找所有群成员
	//var oldGroupMembers []*models.GroupMember
	//_ = db.Where("group_id = ?", req.GroupID).Find(&oldGroupMembers).Error
	//// 添加群成员信息
	//var this models.UserInfoViewer
	//// NOTE 这里直接可以从token 获取到用户信息
	////err = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&this).Error
	//// 正确的写法
	//err = db.Where("user_id = ?", tokenInfo.GetUserID()).Take(&this).Error
	//if err != nil {
	//	logger.Error("查询用户失败", zap.Int64("user_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
	//	return
	//}
	//thisGroupMember := models.GroupMember{
	//	GroupID:  group.GroupID,
	//	MemberID: this.UserID,
	//	//UserType: codes.GroupUserTypeNormal,
	//	Nick: this.Nick,
	//}
	//db.Create(&thisGroupMember)
	//
	//// TODO 向所有成员推送"新人入群"通知
	//if len(oldGroupMembers) != 0 {
	//
	//}
	//// TODO 推送当前用户新聊天事件

	//NOTE  重写 业务不清晰

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

	var currentUserInfo models.UserInfoViewer
	// 正确的写法
	err = db.Where("user_id = ?", tokenInfo.GetUserID()).Take(&currentUserInfo).Error
	if err != nil {
		logger.Error("查询用户失败", zap.Int64("user_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
		return
	}

	// TODO 群组有权限的还要进行判断
	thisGroupMember := models.GroupMember{
		GroupID:  group.GroupID,
		MemberID: currentUserInfo.UserID,
		//UserType: codes.GroupUserTypeNormal,
		Nick: currentUserInfo.Nick,
	}
	// 涉及到数据修改的 一定要做错误处理
	addErr := db.Create(&thisGroupMember).Error
	if addErr != nil {
		logger.Error("添加到群组失败", zap.Error(addErr))
		err = errors.New("加入群组失败")
		return
	}

	go func() {
		// 开启一个线程推送
		// 这个比较简单 只有一个用户 直接推送生成 推送即可
		newGroupEvent := &api.UpdateGroupNewMemberDataType{
			UpdatedAt: thisGroupMember.UpdatedAt,
			//MemberNick: groupInfo.Nick,
			InvitedBy: tokenInfo.GetUserID(),
			MemberID:  thisGroupMember.MemberID,
			//MessageID: msg.ID,
		}

		pushMessage := genMessageAddMemberMessageToDB(p, tokenInfo, group, newGroupEvent)

		// 将消息推送给群里的所有人
		pushGroupMemberMessageFunc(p, currentGroupMembers, pushMessage)

	}()

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

// GetChatIDByBaseGroupID 生成群ID（规则是群ID的负数）
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
	//TODO 查询不到群用户信息
	//var thisUserInfoViewer models.UserInfoViewer
	////_ = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&thisUserInfoViewer).Error
	//_ = db.Where("user_id = ?", tokenInfo.GetUserID()).Take(&thisUserInfoViewer).Error
	////if thisUserInfoViewer.UserType == codes.GroupUserTypeNormal
	//// 不要使用魔鬼数字
	//if thisUserInfoViewer.UserType != int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeAdmin) {
	//	logger.Info("用户无权限增加群通知", zap.Int64("user_id", thisUserInfoViewer.UserID))
	//	return
	//}

	var myInfoByGroup models.GroupMember

	findMemberErr := db.Model(&myInfoByGroup).Where(&models.GroupMember{
		GroupID:  req.GroupID,
		MemberID: tokenInfo.GetUserID(),
	}).Find(&myInfoByGroup).Error

	if findMemberErr != nil || myInfoByGroup.MemberID == 0 || myInfoByGroup.MemberID != tokenInfo.GetUserID() {
		logger.Info("校验群成员信息失败", zap.Error(findMemberErr))
		err = errors.New("邀请入群失败,没有权限操作")
		return
	}

	if myInfoByGroup.UserType != int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeAdmin) {
		logger.Info("用户无权限增加群通知", zap.Int64("user_id", myInfoByGroup.MemberID))

		err = errors.New("没有权限操作")
		return
	}

	// 查找群
	var group models.GroupBaseInfo
	err = db.Where("group_id = ?", req.GroupID).Take(&group).Error
	// 失败，return
	if err != nil {
		logger.Error("查无此群", zap.Int64("group_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
		return
	}
	// 向所有用户推送通知
	// 增加一条消息 到数据库

	// 读取群成员

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

	// 判断我在不在群里

	var memberMySelf *models.GroupMember
	for _, member := range currentGroupMembers {
		if member.MemberID == tokenInfo.GetUserID() {
			memberMySelf = member
			break
		}
	}

	if memberMySelf == nil {
		// 我不再群里 不能发通知
		err = errors.New("我不在群里")
		return
	}

	genMsgID := p.GenMsgID()

	newNotificationMessage := &models.SingleMessage{
		MsgID:     genMsgID.Int64(),
		CreatedAt: genMsgID.Time(),
		ChatID:    GetChatIDByBaseGroupID(req.GroupID),
		MsgStatus: int(api.MessageStatusEnum_MessageStatusSuccess),
		MsgType:   int(api.MessageTypeEnum_MessageTypeUpdateGroupNotification),
		Text:      req.Notification,
		Sender:    tokenInfo.GetUserID(),
	}

	// 添加到数据库

	//

	p.svr.saveMessageChan <- newNotificationMessage

	// 发送的消息

	senderMsg := &api.Message{
		ChatID:      newNotificationMessage.ChatID,
		CreatedAt:   newNotificationMessage.CreatedAt,
		Sender:      newNotificationMessage.Sender,
		MessageText: newNotificationMessage.Text,
		Status:      api.MessageStatusEnum_MessageStatusSuccess,
		Type:        api.MessageTypeEnum_MessageTypeUpdateGroupNotification,
		ID:          newNotificationMessage.MsgID,
	}

	go pushGroupMemberMessageFunc(p, currentGroupMembers, senderMsg)
	//var allGroupMembers []*models.GroupMember
	//_ = db.Where("group_id = ?", req.GroupID).Find(&allGroupMembers).Error
	//if len(allGroupMembers) != 0 {
	//	//for _, m := range allGroupMembers {
	//	//
	//	//}
	//}
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
	// TODO 这个方法是查询不到 群组信息的 只能查到 user 表 不能查询到 group member 表
	//var thisUserInfoViewer models.UserInfoViewer
	////_ = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&thisUserInfoViewer).Error
	//_ = db.Where("user_id = ?", tokenInfo.GetUserID()).Take(&thisUserInfoViewer).Error
	////if thisUserInfoViewer.UserType == codes.GroupUserTypeNormal
	//if thisUserInfoViewer.UserType == 0 {
	//	logger.Info("用户无权限删除用户", zap.Int64("user_id", thisUserInfoViewer.UserID))
	//	return
	//}
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

	// 删除事件
	// 删除成员可以不用关系成员存在问题,因为成员只有存在的时候才能删除

	// 查找删除的用户

	// 校验身份
	var myInfoByGroup models.GroupMember

	findMemberErr := db.Model(&myInfoByGroup).Where(&models.GroupMember{
		GroupID:  req.GroupID,
		MemberID: tokenInfo.GetUserID(),
	}).Find(&myInfoByGroup).Error

	if findMemberErr != nil || myInfoByGroup.MemberID == 0 || myInfoByGroup.MemberID != tokenInfo.GetUserID() {
		logger.Info("校验群成员信息失败", zap.Error(findMemberErr))
		err = errors.New("邀请入群失败,没有权限操作")
		return
	}

	if myInfoByGroup.UserType != int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeAdmin) {
		logger.Info("用户无权限增加群通知", zap.Int64("user_id", myInfoByGroup.MemberID))

		err = errors.New("没有权限操作")
		return
	}

	var findDeleteMembers []*models.GroupMember

	findErr := db.Model(&models.GroupMember{}).Where("user_id in ? ", req.Members).Find(&findDeleteMembers).Error

	if findErr != nil || len(findDeleteMembers) == 0 {
		logger.Error("移除成员错误", zap.Error(findErr))
		err = errors.New("移除错误")
		return
	}

	// 需要遍历 一下id

	// 为了减少 推送量 ,api 也匹配数组的方式推送

	deleteUpdateType := &api.UpdateGroupRemoveMemberDataType{}

	for _, member := range findDeleteMembers {
		deleteUpdateType.MemberID = append(deleteUpdateType.MemberID, member.MemberID)
	}

	removeUserErr := db.Where("user_id in ? and group_id = ?", deleteUpdateType.MemberID).Delete(&models.GroupMember{}).Error

	if removeUserErr != nil {
		logger.Error("删除错误信息", zap.Error(removeUserErr))
	}

	// 查找剩余的成员

	var findOtherMembers []*models.GroupMember

	findErr = db.Model(&models.GroupMember{}).Where(&models.GroupMember{
		GroupID: req.GroupID,
	}).Find(&findOtherMembers).Error

	if findErr != nil {
		err = errors.New("查询成员失败, 请重试")
		return
	}

	genMsgID := p.GenMsgID()
	deleteUpdateType.UpdatedAt = genMsgID.Time()
	deleteUpdateType.ManagerID = tokenInfo.GetUserID()
	deleteUpdateType.MessageID = genMsgID.Int64()

	// 保存到数据库

	paramData, _ := anypb.New(deleteUpdateType)
	newNotificationMessage := &models.SingleMessage{
		MsgID:     genMsgID.Int64(),
		CreatedAt: genMsgID.Time(),
		Sender:    tokenInfo.GetUserID(),
		//Params:    paramData,
		MsgStatus: int(api.MessageStatusEnum_MessageStatusSuccess),
		MsgType:   int(api.MessageTypeEnum_MessageTypeRemoveMember),
	}
	//
	newNotificationMessage.Params, _ = json.Marshal(paramData)

	senderMsg := &api.Message{
		ChatID:      newNotificationMessage.ChatID,
		CreatedAt:   newNotificationMessage.CreatedAt,
		Sender:      newNotificationMessage.Sender,
		MessageText: newNotificationMessage.Text,
		Status:      api.MessageStatusEnum_MessageStatusSuccess,
		Type:        api.MessageTypeEnum_MessageTypeRemoveMember,
		ID:          newNotificationMessage.MsgID,
	}

	senderMsg.Params = paramData

	// 推送给成员

	go pushGroupMemberMessageFunc(p, findDeleteMembers, senderMsg)

	resp = new(api.BaseOk)
	return
}

// 向群成员推送消息的方法
func pushGroupMemberMessageFunc(p *PimServer, currentGroupMembers []*models.GroupMember, msg *api.Message) {

	// 这样就可以少推送一条消息
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
		//findUserClient.PushUserEvent(groupNewEvent)

	}

}

// 生成Message并保存到数据库
func genMessageAddMemberMessageToDB(p *PimServer, tokenInfo TokenInfo, group models.GroupBaseInfo, groupInfo *api.UpdateGroupNewMemberDataType) *api.Message {
	genMsgID := p.GenMsgID()
	newMessage := &models.SingleMessage{
		MsgID:     genMsgID.Int64(),
		CreatedAt: genMsgID.Time(),
		ChatID:    GetChatIDByBaseGroupID(group.GroupID), // 注意 群的规则是 负号
		Sender:    tokenInfo.GetUserID(),
		MsgType:   int(api.MessageTypeEnum_MessageTypeNewMember),
		MsgStatus: int(api.MessageStatusEnum_MessageStatusSend),
	}

	// 顺便把msgID给改了
	groupInfo.MessageID = newMessage.MsgID
	paramsData, _ := anypb.New(groupInfo)

	newMessage.Params, _ = json.Marshal(paramsData)

	pushMessage := &api.Message{}
	pushMessage.Sender = newMessage.Sender
	pushMessage.ChatID = newMessage.ChatID
	pushMessage.ID = newMessage.MsgID
	pushMessage.Type = api.MessageTypeEnum_MessageTypeNewMember
	pushMessage.Status = api.MessageStatusEnum_MessageStatusSend
	pushMessage.Params = paramsData

	// saveMessageChan就是丢数据库里
	p.svr.saveMessageChan <- newMessage
	//p.svr.sendMessageChan <- pushMessage

	return pushMessage
}

// 这是耗时的操作
func runGroupMemberInviteProc(p *PimServer, tokenInfo TokenInfo, group models.GroupBaseInfo, currentGroupMembers []*models.GroupMember, checkMemberUserInfo []*models.UserInfoViewer) {

	// 批量推送消息 循环推送
	// 参照微信 , 先入的能够收到后面的消息

	// 协议个推送的方法

	// 遍历邀请的用户信息
	db := p.svr.db
	logger := p.svr.logger
	// 用于生成一条需要发送的消息 其中一条是存数据库的

	for _, userViewer := range checkMemberUserInfo {
		// 	按照顺序生成 消息

		// 生成成员信息

		newMemberInfo := &models.GroupMember{
			GroupID:  group.GroupID,
			MemberID: userViewer.UserID,
			Nick:     userViewer.Nick,
			Inviter:  tokenInfo.GetUserID(), // 增加一个 邀请人 业务会更完整 , 还有进群方式类型等...
			UserType: int(api.GroupMemberUserEnumType_GroupMemberUserEnumTypeNormal),
		}

		// 添加到群表
		addGroupErr := db.Create(newMemberInfo).Error
		if addGroupErr != nil {
			// 添加失败
			logger.Error("添加成员信息失败", zap.Error(addGroupErr))
			continue
		}

		// 添加成功推送一个 入群事件
		newGroupEvent := &api.UpdateGroupNewMemberDataType{
			UpdatedAt: newMemberInfo.UpdatedAt,
			InvitedBy: tokenInfo.GetUserID(),
			MemberID:  newMemberInfo.MemberID,
		}

		pushMessage := genMessageAddMemberMessageToDB(p, tokenInfo, group, newGroupEvent)

		// 将消息推送给群里的所有人
		pushGroupMemberMessageFunc(p, currentGroupMembers, pushMessage)
		// 推玩就要吧当前用户加到列表
		// 下一轮则也会推送这个用户的数据
		currentGroupMembers = append(currentGroupMembers, newMemberInfo)

	}
}

// GroupsCache groupID -> userID -> groupMember
type GroupsCache map[int64]*SingleGroupCache
type SingleGroupCache map[int64]*models.GroupMember

func (g GroupsCache) PushEventToGroup(p *PimServer, groupID int64, event *api.UpdateEventDataType) (single *SingleGroupCache, err error) {
	// TODO implements
	group, has := g[groupID]
	if has {
		group.PushEventsToEveryone(p, event)
	}
	return
}

func (s SingleGroupCache) PushEventsToEveryone(p *PimServer, event *api.UpdateEventDataType) {
	// TODO implements
	for _, member := range s {
		p.UserStreamClientMap.PushUserEvent(member.MemberID, event)
	}
	return
}

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
