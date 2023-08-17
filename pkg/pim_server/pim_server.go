package pim_server

import (
	"errors"
	"pim/api"
	"pim/pkg/tools"
	"sync"
)

type RpcClient struct {
	//	一个链接的客户端
	UserID   int64
	Pf       int
	StreamID int64
	Level    int
	PushFunc func(event *api.UpdateEventDataType)
	svr      *server
}

// GetUserID 实现token 的接口
func (c *RpcClient) GetUserID() int64 {
	return c.UserID
}

func (c *RpcClient) GetPf() int {
	return c.Pf
}

func (c *RpcClient) GetLevel() int {
	return c.Level
}

type StreamInfoReq interface {
	GetStreamID() int64
}

// CheckAuthByStream 通过流id 鉴权
func (p *PimServer) CheckAuthByStream(req StreamInfoReq) (token TokenInfo, err error) {
	p.rw.RLock()
	defer p.rw.RUnlock()

	client, isok := p.clients[req.GetStreamID()]
	if !isok {
		err = errors.New("没有权限")
		return
	}

	token = client

	return
}

func (p *PimServer) GenMsgID() tools.ID {
	return p.svr.msgNode.Generate()
}

type PimServer struct {
	svr *server
	// 这个map 是调用接口的时候快速查询用的
	// 使用 流id 查询基本信息
	clients map[int64]*RpcClient
	rw      *sync.RWMutex
	// 使用 用户id 查找流信息
	UserStreamClientMap UserStreamClientMapType
	groups              map[int64][]int64
}

func SetNodeID() Option {
	return func(svr *server) {
		svr.msgNode, _ = tools.NewNode(int64(1))
	}
}
