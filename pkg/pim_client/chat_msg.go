package pim_client

import (
	"github.com/jroimartin/gocui"
)

type ChatMsgWidget struct {
	BasePos *BaseUIArea
	pos     *TargetPos
	client  *PimClient
}

func (w *ChatMsgWidget) Layout(g *gocui.Gui) error {
	//TODO implement me
	//panic("implement me")

	pos := w.pos

	if pos.Hide {
		return nil
	}

	if v, err := g.SetView(pos.Title, pos.StartX, pos.StartY, pos.StartX+pos.StartWidth, pos.StartY+pos.StartHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = pos.Title
		v.Autoscroll = true
		v.Wrap = true

		//v.Autoscroll = true

		//w.ShowList(v)
	}

	return nil
}
