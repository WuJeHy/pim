package pim_server

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"pim/api"
	"sync"
)

type RpcClient struct {
	//	一个链接的客户端
	PushFunc func(event *api.UpdateEventDataType)
}

type PimServer struct {
	svr *server
	// 这个map 是调用接口的时候快速查询用的
	clients map[int64]*RpcClient
	rw      *sync.RWMutex
	// 用户映射还没加
}

func (p *PimServer) UpdateEvent(req *api.TokenReq, eventServer api.PimServer_UpdateEventServer) error {
	//TODO implement me
	//panic("implement me")

	streamID := int64(uuid.New().ID())

	logger := p.svr.logger.Named(fmt.Sprintf("%d", streamID))

	eventChannel := make(chan *api.UpdateEventDataType)
	lock := new(sync.Mutex)
	client := &RpcClient{
		PushFunc: func(event *api.UpdateEventDataType) {
			lock.Lock()
			defer lock.Unlock()
			eventChannel <- event
		},
	}

	p.rw.Lock()
	p.clients[streamID] = client
	p.rw.Unlock()

	defer func() {
		p.rw.Lock()
		delete(p.clients, streamID)
		p.rw.Unlock()

	}()

	for true {
		select {
		case <-p.svr.closeServer:
			logger.Info("主服务发出退出指令")
			return nil

		case <-eventServer.Context().Done():
			logger.Info("Exit ctx", zap.Error(eventServer.Context().Err()))
			return eventServer.Context().Err()

		case pushEvent := <-eventChannel:

			err := eventServer.Send(pushEvent)

			if err != nil {
				logger.Error("send err", zap.Error(err))
				return err
			}

		}
	}

	return nil
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
