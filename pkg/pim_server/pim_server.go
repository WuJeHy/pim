package pim_server

import (
	"context"
	"pim/api"
)

type PimServer struct {
	svr *server
}

func (p *PimServer) UpdateEvent(ctx context.Context, req *api.TokenReq) (resp *api.UpdateEventDataType, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) Register(ctx context.Context, req *api.RegisterReq) (rtesp *api.BaseOk, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) Login(ctx context.Context, req *api.LoginReq) (resp *api.LoginResp, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) GetMyUserInfo(ctx context.Context, req *api.StreamReq) (resp *api.UserInfoViewerDataType, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) GetUserInfoByID(ctx context.Context, req *api.GetUserInfoByIDReq) (resp *api.UserInfoViewerDataType, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) AddUserToContact(ctx context.Context, req *api.AddUserToContactReq) (resp *api.BaseOk, err error) {
	//TODO implement me
	panic("implement me")
}

func (p *PimServer) SendMessage(ctx context.Context, req *api.SendMessageReq) (resp *api.SendMessageResp, err error) {
	//TODO implement me
	panic("implement me")
}
