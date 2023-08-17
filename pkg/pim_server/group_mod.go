package pim_server

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"pim/api"
	"pim/pkg/models"
	time "time"
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
			UpdateAt:  groupCreatedTime,
			InvitedBy: groupMasterInfo.Nick,
		}

		for _, m := range membersInfo {
			temp := &models.GroupMember{
				GroupID:  groupBaseInfo.GroupID,
				MemberID: m.MemberID,
				Nick:     m.Nick,
				//UserType: codes.GroupUserTypeNormal,
			}
			groupMemberList = append(groupMemberList, temp)
			p.pushCacheToGroups(groupBaseInfo.GroupID, temp.MemberID)

			updateGroupNewMemberDataTypeModel.NewMemberNick = temp.Nick
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
	newChatInfoDataType := api.NewChatInfoDataType{
		ChatName:  groupBaseInfo.Name,
		ChatTitle: groupBaseInfo.Name,
		ChatTo:    groupBaseInfo.GroupID,
		UpdateAt:  groupCreatedTime,
	}
	newChatInfoBody, _ := anypb.New(&newChatInfoDataType)
	newChatInfoPushedData := &api.UpdateEventDataType{
		Type: api.UpdateEventDataType_NewChatInfo,
		Body: newChatInfoBody,
	}

	for _, m := range req.Members {
		p.UserStreamClientMap.PushUserEvent(m, newChatInfoPushedData)
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
	// 查找所有群成员
	var oldGroupMembers []*models.GroupMember
	_ = db.Where("group_id = ?", req.GroupID).Find(&oldGroupMembers).Error
	// 添加群成员信息
	var this models.UserInfoViewer
	err = db.Where("user_id = ?", p.clients[req.StreamID].UserID).Take(&this).Error
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
	var thisMembers []*models.UserInfoViewer
	err = db.Where("user_id in ?", req.Members).Find(&thisMembers).Error
	if err != nil {
		logger.Error("添加群成员信息失败", zap.Int64("group_id", req.GroupID), zap.Int64("stream_id", req.StreamID))
		return
	}
	var groupMembers []*models.GroupMember
	if len(thisMembers) != 0 {
		for _, am := range thisMembers {
			temp := models.GroupMember{
				GroupID:  group.GroupID,
				MemberID: am.UserID,
				Nick:     am.Nick,
				//UserType: codes.GroupUserTypeNormal,
			}
			groupMembers = append(groupMembers, &temp)
		}
	}
	_ = db.Create(&groupMembers)
	// TODO 向新用户们推送新聊天事件
	// 查找所有群成员
	var allGroupMembers []*models.GroupMember
	_ = db.Where("group_id = ?", req.GroupID).Find(&allGroupMembers).Error
	// TODO 向所有成员推送"新人入群"通知
	if len(allGroupMembers) != 0 {
		//for _, m := range groupMembers{
		//}
	}
	resp = new(api.BaseOk)
	return
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
	var deletedGroupMembers []models.GroupMember
	for k, m := range req.Members {
		deletedGroupMembers[k] = models.GroupMember{
			MemberID: m,
		}
	}
	_ = db.Delete(&deletedGroupMembers).Error
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
	return p.clients[streamID].UserID
}
