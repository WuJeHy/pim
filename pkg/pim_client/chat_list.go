package pim_client

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"pim/api"
	"sort"
)

type ChatListWidget struct {
	BasePos         *BaseUIArea
	pos             *TargetPos
	currentTopIndex int
	chatMap         map[int64]*api.ChatInfoDataType
	client          *PimClient
	testList        []string
	testIndex       int
}

func (w *ChatListWidget) Bind(gui *gocui.Gui) error {
	//TODO implement me
	//panic("implement me")

	pos := w.pos
	if err := gui.SetKeybinding(pos.Title, gocui.MouseWheelDown, gocui.ModNone, w.UpdataListDown(pos)); err != nil {
		return err
	}
	if err := gui.SetKeybinding(pos.Title, gocui.MouseWheelUp, gocui.ModNone, w.UpdataListUp(pos)); err != nil {
		return err
	}
	return nil
}

func NewChatListWidget(client *PimClient, base *BaseUIArea) *ChatListWidget {

	var testList []string

	for i := 0; i < 120; i++ {
		testList = append(testList, fmt.Sprintf("List_%d", i))
	}

	return &ChatListWidget{
		BasePos:  base,
		pos:      base.GetChatListPos(),
		client:   client,
		testList: testList,
	}
}

func (w *ChatListWidget) Layout(g *gocui.Gui) error {
	// 绘制ui 的方法
	// 这里绘制的是列表

	// 从最表 x y 开始显示

	pos := w.pos

	if pos.Hide {
		return nil
	}

	if v, err := g.SetView(pos.Title, pos.StartX, pos.StartY, pos.StartX+pos.StartWidth, pos.StartY+pos.StartHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = pos.Title
		//v.Autoscroll = true

		w.ShowList(v)
	}

	return nil
}

func (w *ChatListWidget) UpdataListUp(pos *TargetPos) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		if view.Title != pos.Title {
			//fmt.Println("ui fail", view.Title)
			return nil
		}

		if w.testIndex > 0 {
			w.testIndex--
		}

		view.Clear()
		w.ShowList(view)

		return nil
	}
}

func (w *ChatListWidget) UpdataListDown(pos *TargetPos) func(*gocui.Gui, *gocui.View) error {
	return func(gui *gocui.Gui, view *gocui.View) error {
		if view.Title != pos.Title {
			//fmt.Println("ui fail", view.Title)
			return nil
		}

		if w.testIndex+pos.StartHeight < len(w.client.ChatInfos) {
			w.testIndex++
		}
		view.Clear()

		w.ShowList(view)

		return nil
	}
}

func (w *ChatListWidget) ShowList(view *gocui.View) {

	pos := w.pos

	// 读取信息
	w.chatMap = w.client.ChatInfos

	var chatListUpdata []*api.ChatInfoDataType

	if len(w.chatMap) == 0 {
		return
	}

	for _, dataType := range w.chatMap {
		chatListUpdata = append(chatListUpdata, dataType)
	}

	sort.SliceIsSorted(chatListUpdata, func(i, j int) bool {
		return chatListUpdata[i].LastUpdateTime > chatListUpdata[j].LastUpdateTime
	})

	var showListMax = pos.StartHeight
	if pos.StartHeight >= len(chatListUpdata)-w.testIndex {
		showListMax = len(w.chatMap)
	}

	//fmt.Fprintln(view, "")
	for i := 0; i < showListMax; i++ {
		current := chatListUpdata[w.testIndex+i]
		showTitle := current.ChatTitle
		if showTitle == "" {
			showTitle = current.ChatName
		}
		fmt.Fprintln(view, showTitle)
	}
}
