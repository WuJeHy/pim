package pim_server

import (
	"context"
	"go.uber.org/zap"
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
	// 插入群成员信息
	var membersInfo []*models.GroupMember
	_ = db.Where("user_id in ?", req.Members).Find(&membersInfo).Error

	if len(membersInfo) != 0 {
		var groupMemberList []*models.GroupMember
		for _, m := range membersInfo {
			temp := &models.GroupMember{
				GroupID:  groupBaseInfo.GroupID,
				MemberID: m.MemberID,
				Nick:     m.Nick,
				//UserType: codes.GroupUserTypeNormal,
			}
			groupMemberList = append(groupMemberList, temp)
		}
		db.Create(groupMemberList)
	}
	resp = new(api.CreateGroupResp)
	// TODO 向群成员推送新聊天事件
	// TODO 向群成员推送"欢迎加入"事件
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
	// 查找所有群成员
	var groupMembers []*models.GroupMember
	_ = db.Where("group_id = ?", req.GroupID).Find(&groupMembers).Error
	// TODO 向所有成员推送"新人入群"通知
	if len(groupMembers) != 0 {
		//for _, m := range groupMembers{
		//}
	}
	// TODO 推送当前用户新聊天事件
	resp = new(api.BaseOk)
	return
}

func (p *PimServer) GroupInviteMembers(ctx context.Context, req *api.GroupInviteMembersReq) (resp *api.BaseOk, err error) {
	// TODO 鉴权
	// 鉴权失败
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

	return
}

func (p *PimServer) GroupRemoveMembers(ctx context.Context, req *api.GroupRemoveMembersReq) (resp *api.BaseOk, err error) {
	//TODO implement me
	return
}
