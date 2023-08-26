package pim_client

import "github.com/jroimartin/gocui"

type TargetPos struct {
	Title       string
	StartX      int
	StartY      int
	StartWidth  int
	StartHeight int
	Hide        bool
}

// BaseUIArea 基础ui 分区指标
type BaseUIArea struct {
	// 聊天框
	ChatInfo TargetPos
	ChatList TargetPos
	// 我的信息坐标
	MyInfo TargetPos
	//
	ChatMsg   TargetPos
	ChatInput TargetPos
	ChatSend  TargetPos
}

func (b *BaseUIArea) Bind(gui *gocui.Gui) error {
	//TODO implement me
	//panic("implement me")
	return nil
}

func (b *BaseUIArea) GetChatInfoPos() *TargetPos {
	return &b.ChatInfo
}
func (b *BaseUIArea) GetChatListPos() *TargetPos {
	return &b.ChatList
}
func (b *BaseUIArea) Layout(gui *gocui.Gui) error {
	//TODO implement me
	//panic("implement me")

	maxX, maxY := gui.Size()

	_ = maxX

	b.MyInfo.Title = "MyInfo"
	b.MyInfo.StartX = 1
	b.MyInfo.StartY = 1
	b.MyInfo.StartWidth = 20
	b.MyInfo.StartHeight = 5
	//b.MyInfo.Hide = true

	b.ChatList.Title = "ChatList"
	b.ChatList.StartX = b.MyInfo.StartX
	b.ChatList.StartY = b.MyInfo.StartY + b.MyInfo.StartHeight

	b.ChatList.StartWidth = b.MyInfo.StartWidth

	b.ChatList.StartHeight = maxY - 3 - b.MyInfo.StartHeight

	b.ChatMsg.Title = "Msg"
	b.ChatMsg.StartX = b.MyInfo.StartX + b.MyInfo.StartWidth + 1
	b.ChatMsg.StartY = b.MyInfo.StartY
	b.ChatMsg.StartWidth = maxX - b.ChatMsg.StartX - 15
	b.ChatMsg.StartHeight = maxY - 2 - 5

	b.ChatSend.Title = "Send"
	b.ChatSend.StartX = b.ChatMsg.StartX
	b.ChatSend.StartY = b.ChatMsg.StartY + b.ChatMsg.StartHeight + 1
	b.ChatSend.StartWidth = maxX - b.ChatSend.StartX - 15
	b.ChatSend.StartHeight = 2

	b.ChatInfo.Title = "ChatInfo"
	b.ChatInfo.StartX = b.ChatMsg.StartX + b.ChatMsg.StartWidth + 1
	b.ChatInfo.StartY = b.ChatMsg.StartY
	b.ChatInfo.StartWidth = maxX - b.ChatInfo.StartX - 1
	b.ChatInfo.StartHeight = b.ChatMsg.StartHeight

	return nil
}

func (b *BaseUIArea) GetMyInfoPos() *TargetPos {
	return &b.MyInfo
}

func (b *BaseUIArea) GetChatMsgPos() *TargetPos {
	return &b.ChatMsg
}

func (b *BaseUIArea) GetChatSendPos() *TargetPos {
	return &b.ChatSend
}
