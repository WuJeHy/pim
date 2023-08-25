package pim_client

import (
	"github.com/jroimartin/gocui"
	"go.uber.org/zap"
	"strings"
)

type ChatSendWidget struct {
	BasePos *BaseUIArea
	pos     *TargetPos
	client  *PimClient
}

func (c *ChatSendWidget) Bind(gui *gocui.Gui) error {
	pos := c.BasePos.GetChatSendPos()
	if err := gui.SetKeybinding(pos.Title, gocui.MouseRelease, gocui.ModNone, c.ProcLeftButton(pos)); err != nil {
		return err
	}
	if err := gui.SetKeybinding(pos.Title, gocui.KeyEnter, gocui.ModNone, c.SenderMessage(pos)); err != nil {
		return err
	}

	return nil
}

func (c *ChatSendWidget) Layout(g *gocui.Gui) error {
	//TODO implement me
	pos := c.pos
	if v, err := g.SetView(pos.Title, pos.StartX, pos.StartY, pos.StartX+pos.StartWidth, pos.StartY+pos.StartHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = pos.Title
		v.Autoscroll = true
		v.Wrap = true
		v.Editable = true

		//v.Autoscroll = true

		//w.ShowList(v)
	}
	return nil
}

func (c *ChatSendWidget) ProcLeftButton(pos *TargetPos) func(*gocui.Gui, *gocui.View) error {
	//return func(gui *gocui.Gui, view *gocui.View) error {
	//	//view.Editable = true
	//	//err :=
	//	//
	//	//if pos.Title {
	//	//
	//	//}
	//
	//	if _, err := gui.SetCurrentView(pos.Title); err != nil {
	//		return err
	//	}
	//
	//	//
	//
	//	data := view.ViewBuffer()
	//
	//	toSpace := strings.TrimSpace(data)
	//	if len(toSpace) == 0 {
	//		view.SetCursor(0, 0)
	//
	//		return nil
	//	}
	//
	//	//x, y := view.Cursor()
	//	//fmt.Println(x, y)
	//	dataList := view.ViewBufferLines()
	//	line := len(dataList)
	//
	//	//dataToSpace := strings.TrimSpace(data)
	//
	//	var x = 0
	//	if line > 0 {
	//		toSpace := strings.TrimSpace(dataList[line-1])
	//		x = len(toSpace)
	//	}
	//	//
	//	//if _, err := gui.SetViewOnTop(pos.Title); err != nil {
	//	//
	//	//	return err
	//	//}
	//	return view.SetCursor(x, line)
	//}
	return cursorDown
}

func (c *ChatSendWidget) SenderMessage(pos *TargetPos) func(*gocui.Gui, *gocui.View) error {
	logger := c.client.logger
	return func(gui *gocui.Gui, view *gocui.View) error {
		msg := strings.TrimSpace(view.ViewBuffer())
		logger.Info("send msg", zap.String("msg", msg))

		view.EditNewLine()
		view.Clear()
		return nil
	}
}
func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		if _, err := g.SetCurrentView(v.Title); err != nil {
			return err
		}
		//cx, cy := v.Cursor()
		viewData := v.ViewBuffer()

		if err := v.SetCursor(0, len(strings.TrimSpace(viewData))); err != nil {
			//ox, oy := v.Origin()
			//if err := v.SetOrigin(ox, oy); err != nil {
			//	return err
			//}
		}
	}
	return nil
}
func NewChatSendWidget(c *PimClient, base *BaseUIArea) *ChatSendWidget {
	return &ChatSendWidget{
		client:  c,
		BasePos: base,
		pos:     base.GetChatSendPos(),
	}
}
