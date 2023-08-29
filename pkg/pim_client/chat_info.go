package pim_client

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"pim/api"
	"time"
)

type ChatInfoWidget struct {
	BasePos         *BaseUIArea
	pos             *TargetPos
	client          *PimClient
	currentChatInfo *api.ChatInfoDataType
}

func (c *ChatInfoWidget) Layout(gui *gocui.Gui) error {
	pos := c.BasePos.GetChatInfoPos()

	if pos.Hide {
		return nil
	}

	if v, err := gui.SetView(pos.Title, pos.StartX, pos.StartY, pos.StartX+pos.StartWidth, pos.StartY+pos.StartHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = pos.Title
		v.Autoscroll = true
		v.Wrap = true
		//v.Highlight = true
		//v.SelBgColor = gocui.ColorGreen
		//v.SelFgColor = gocui.ColorBlack
		//fmt.Fprintln(v, "AAA")
		//fmt.Fprintln(v, "BBB")
		//fmt.Fprintln(v, "CCC")
		if c.currentChatInfo == nil {
			return nil
		}

		info := c.currentChatInfo
		fmt.Fprintln(v, info.ChatId)
	}

	return nil
}

func (c *ChatInfoWidget) updataView(gui *gocui.Gui) error {
	view, err := gui.View(c.pos.Title)
	if err != nil {
		return nil
	}

	if c.currentChatInfo == nil {
		return nil
	}

	info := c.currentChatInfo
	view.EditNewLine()
	view.Clear()
	fmt.Fprintln(view, "ChatID:")
	fmt.Fprintf(view, "%d", info.ChatId)
	fmt.Fprintln(view)
	if info.ChatTitle != "" {
		fmt.Fprintln(view, "ChatName:")
		fmt.Fprintf(view, "%s", info.ChatTitle)
		fmt.Fprintln(view)
	} else {
		if info.ChatName != "" {
			fmt.Fprintln(view, "ChatName:")
			fmt.Fprintf(view, "%s", info.ChatName)
			fmt.Fprintln(view)
		}
	}

	fmt.Fprintln(view, "LastMsg:")
	fmt.Fprintf(view, "%d\n", info.LastMsgId)
	fmt.Fprintln(view, "LastTime:")
	if info.LastUpdateTime == 0 {
		return nil
	}
	timeAt := time.UnixMilli(info.LastUpdateTime)
	fmt.Fprintf(view, "%s\n", timeAt.Format(time.RFC3339))

	return nil
}

func NewChatInfoWidget(client *PimClient, base *BaseUIArea) *ChatInfoWidget {

	return &ChatInfoWidget{
		BasePos: base,
		pos:     base.GetChatInfoPos(),
		client:  client,
	}
}
