package pim_client

import (
	"fmt"
	"github.com/jroimartin/gocui"
)

type MyInfoWidget struct {
	BasePos *BaseUIArea
	pos     *TargetPos
	client  *PimClient
}

func (c *MyInfoWidget) Bind(gui *gocui.Gui) error {
	//TODO implement me
	//panic("implement me")
	pos := c.BasePos.GetMyInfoPos()
	if err := gui.SetKeybinding(pos.Title, gocui.MouseLeft, gocui.ModNone, c.ProcLeftButton(pos)); err != nil {
		return err
	}

	return nil
}

func (c *MyInfoWidget) Layout(gui *gocui.Gui) error {

	pos := c.BasePos.GetMyInfoPos()

	if pos.Hide {
		return nil
	}

	if v, err := gui.SetView(pos.Title, pos.StartX, pos.StartY, pos.StartX+pos.StartWidth, pos.StartY+pos.StartHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = pos.Title
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		fmt.Fprintln(v, "AAA")
		fmt.Fprintln(v, "BBB")
		fmt.Fprintln(v, "CCC")

	}

	return nil
}

func (c *MyInfoWidget) ProcLeftButton(pos *TargetPos) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		//var l string
		//var err error
		//
		//if _, err := gui.SetCurrentView(view.Name()); err != nil {
		//	return err
		//}
		//
		//if view.Title != c.BasePos.GetMyInfoPos().Title {
		//	//fmt.Println("不是目标ui", view.Title)
		//
		//	return nil
		//}
		//
		//_, cy := view.Cursor()
		//if l, err = view.Line(cy); err != nil {
		//	l = ""
		//}

		//fmt.Println("current ", l)

		//maxX, maxY := gui.Size()
		//if v, err := gui.SetView("msg", maxX/2-10, maxY/2, maxX/2+10, maxY/2+2); err != nil {
		//	if err != gocui.ErrUnknownView {
		//		return err
		//	}
		//	fmt.Fprintln(v, l)
		//}
		//gui.Update(func(gui *gocui.Gui) error {
		//chaList, err := gui.View(c.BasePos.ChatList.Title)
		//
		//if err != nil {
		//	return nil
		//}
		//
		//chaList.Clear()
		//fmt.Fprintln(chaList, l)
		//return nil
		//})
		return nil
	}

}
