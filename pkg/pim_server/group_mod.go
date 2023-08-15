package pim_server

import (
	"context"
	"go.uber.org/zap"
	"pim/api"
	"pim/pkg/models"
)

func (p *PimServer) CreateGroup(ctx context.Context, req *api.CreateGroupReq) (resp *api.CreateGroupResp, err error) {
	// TODO 鉴权
	// 鉴权失败，返回群ID为0
	// 鉴权成功
	db := p.svr.db
	logger := p.svr.logger
	// 插入群信息
	groupBaseInfo := &models.GroupBaseInfo{
		Name: req.Name,
	}
	// 创建成功后，主键的值将会被插入groupBaseInfo中
	err = db.Create(groupBaseInfo).Error
	if err != nil {
		logger.Error("群信息插入失败", zap.Int64("stream_id", req.StreamID))
		return
	}
	// 插入群成员信息
	var membersInfo []*models.GroupMember
	_ = db.Model(&models.UserInfoViewer{}).Where("user_id in ?", req.Members).Find(&membersInfo).Error

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
	// TODO 向群成员推送新聊天事件（群主是群成员之一）
	// TODO 向群成员推送"欢迎加入"事件
	return
}

func (p *PimServer) GroupJoinByID(ctx context.Context, req *api.GroupJoinByIDReq) (resp *api.BaseOk, err error) {
	//todo

	return
}

func (p *PimServer) GroupInviteMembers(ctx context.Context, req *api.GroupInviteMembersReq) (resp *api.BaseOk, err error) {
	//todo

	return
}

func (p *PimServer) GroupEditNotification(ctx context.Context, req *api.GroupEditNotificationReq) (resp *api.BaseOk, err error) {
	//todo

	return
}
