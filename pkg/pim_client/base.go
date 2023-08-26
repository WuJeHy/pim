package pim_client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jroimartin/gocui"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"pim/api"
	"pim/pkg/tools"
	"strings"
	"time"
)

type PimClient struct {
	clientApi api.PimServerClient
	logger    *zap.Logger
	//db        *gorm.DB
	ctx                   context.Context
	currentToken          string
	sessionFile           string
	rpcServerUrl          string
	connectStatus         chan bool
	ChatInfos             map[int64]*api.ChatInfoDataType
	currentChatInfoWidget *ChatInfoWidget
	streamID              int64
}

func (c *PimClient) CheckLogin() bool {

	// 判断token 是否有效

	getTokenBytes, err := os.ReadFile(c.sessionFile)

	if err != nil {
		fmt.Println("session 文件不存在,准备进行登录操作")
		return c.doLoginRpc()

	}
	// 校验
	if len(getTokenBytes) == 0 {
		fmt.Println("读取session 错误,准备重新登录 ")
		return c.doLoginRpc()
	}

	fmt.Println("session 读取成功. 尝试登录")

	c.currentToken = string(getTokenBytes)
	return true
}

func (c *PimClient) CheckRpc() bool {

	if c.clientApi == nil {

		//链接rpc
		rpcClient, err := grpc.Dial(c.rpcServerUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			fmt.Println("链接失败", err)
			time.Sleep(time.Second * 5)
			return false
		}

		c.clientApi = api.NewPimServerClient(rpcClient)
	}

	return true
}

func (c *PimClient) doLoginRpc() bool {

	//var phoneStr string

	readerStdin := bufio.NewReader(os.Stdin)

	//  读取一行

	var username string
	var password string

	for i := 0; i < 3; i++ {
		fmt.Print("请输入username:")
		readUsername, err := readerStdin.ReadString('\n')

		if err != nil {
			fmt.Println("输入有误,请重试...")
			continue
		}

		username = strings.TrimSpace(readUsername)
		if username == "" {
			fmt.Println("输入有误,请重试...")
			continue
		}
		break
	}

	fmt.Println("输入成功[", username, "]")

	for i := 0; i < 2; i++ {
		fmt.Print("输入密码:")
		readString, err := readerStdin.ReadString('\n')

		if err != nil {
			fmt.Println("输入有误,请重试...")
			continue
		}

		password = strings.TrimSpace(readString)
		if password == "" {
			fmt.Println("输入有误,请重试...")
			continue
		}
		break
	}

	// 读取token

	loginReq := &api.LoginReq{
		Type:     api.LoginReq_LoginByUsername,
		Platform: api.LoginReq_Grpc,
		Username: username,
		Password: password,
	}

	loginTokenResp, err := c.clientApi.Login(c.ctx, loginReq)

	if err != nil {
		fmt.Println("登录失败", err)
		return false
	}

	// 登录成功 写入token

	os.WriteFile(c.sessionFile, []byte(loginTokenResp.Token), os.ModePerm)

	//readUsername , err :=

	//n, err := fmt.Scanln(&phoneStr)
	// 链接rpc 事件

	return true
}

func NewMyInfoWidget(c *PimClient, ui *BaseUIArea) *MyInfoWidget {
	return &MyInfoWidget{
		BasePos: ui,
		pos:     ui.GetMyInfoPos(),
		client:  c,
	}
}
func (c *PimClient) Run() bool {
	// 打开ui

	fmt.Println("打开ui")
	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		fmt.Println("ui 打开错误", err)
		return false
	}

	_ = g
	defer g.Close()

	go c.runEvent(g)
	// 需要轮循事件
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	g.Mouse = true
	appUI := &BaseUIArea{}

	myInfoWidget := NewMyInfoWidget(c, appUI)
	chatListWidget := NewChatListWidget(c, appUI)
	chatMsgWidget := NewChatMsgWidget(c, appUI)
	chatSendWidget := NewChatSendWidget(c, appUI)
	chatInfoWidget := NewChatInfoWidget(c, appUI)

	c.currentChatInfoWidget = chatInfoWidget

	g.SetManager(appUI, myInfoWidget, chatListWidget, chatMsgWidget, chatSendWidget, chatInfoWidget)
	//g.SetManagerFunc(func(gui *gocui , .Gui) error {
	//	return layout(c, g)
	//})

	if err := keyBindings(g, myInfoWidget, chatListWidget, chatSendWidget); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}

	return true
}

type BindKeyFunc interface {
	Bind(gui *gocui.Gui) error
}

func keyBindings(g *gocui.Gui, ui ...BindKeyFunc) error {
	for _, keyFunc := range ui {
		if err := keyFunc.Bind(g); err != nil {
			return err
		}
	}
	return nil
}

func (c *PimClient) CheckEvent() bool {

	//tokenReq := &api.TokenReq{
	//	Token: c.currentToken,
	//}

	return true
}

func (c *PimClient) runEvent(g *gocui.Gui) {
	tokenReq := &api.TokenReq{
		Token: c.currentToken,
	}
	updateEvent, err := c.clientApi.UpdateEvent(c.ctx, tokenReq)
	if err != nil {
		c.logger.Info("update event fail ", zap.Error(err))
		return
	}

	for true {
		readEvent, errEvent := updateEvent.Recv()
		if errEvent != nil {
			c.logger.Info("收到事件错误", zap.Error(errEvent))
			return
		}

		g.Update(func(gui *gocui.Gui) error {
			findMsg, err := g.View("Msg")

			if err != nil {
				c.logger.Info("not found msg s")
				return err
			}

			var output interface{}
			switch readEvent.Type {
			case api.UpdateEventDataType_NewMessage:
				var newMsg api.Message
				err = readEvent.Body.UnmarshalTo(&newMsg)
				output = &newMsg
			case api.UpdateEventDataType_NewChatInfo:

				var data api.ChatInfoDataType
				err = readEvent.Body.UnmarshalTo(&data)
				output = &data
				c.ChatInfos[data.ChatId] = &data

			case api.UpdateEventDataType_ConnectSuccess:
				var data api.ConnectSuccessDataType
				err = readEvent.Body.UnmarshalTo(&data)
				output = &data
				c.streamID = data.StreamID

			case api.UpdateEventDataType_UpdateUserInfo:

				var data api.UserInfoViewerDataType
				err = readEvent.Body.UnmarshalTo(&data)
				output = &data
			}
			toJson, _ := json.Marshal(output)
			fmt.Fprintf(findMsg, "[%s]:%s\n", readEvent.Type.String(), toJson)

			return nil
		})

	}
}
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func layout(client *PimClient, g *gocui.Gui) error {

	maxX, maxY := g.Size()

	if v, err := g.SetView("Chats", 0, 0, 19, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Chats"
		v.Wrap = true
		v.Autoscroll = true
		v.Editable = true
	}

	if v, err := g.SetView("Msg", 20, 0, maxX-20, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Msg"
		v.Wrap = true
		v.Autoscroll = true
		v.Editable = true

	}

	if v, err := g.SetView("Info", maxX-19, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Info"
	}

	return nil
}

// RunClient 只有一个参数
func RunClient(rpc string, sessionFile string) {

	// 读取session 文件

	logger_level := zapcore.DebugLevel
	logger := tools.LoggerInitLevelTag("logs", "tui_client", &logger_level)
	client := &PimClient{
		ctx:          context.Background(),
		sessionFile:  sessionFile,
		rpcServerUrl: rpc,
		logger:       logger,
		ChatInfos:    map[int64]*api.ChatInfoDataType{},
	}

	//ticker := time.NewTicker(time.Second * 5)

	for true {

		if !client.CheckRpc() {
			fmt.Println("等待链接...")
			time.Sleep(time.Second * 5)
			fmt.Println("重试链接...")
			continue
		} else {
			fmt.Println("链接成功.")
		}

		fmt.Println("校验登录")
		if !client.CheckLogin() {
			fmt.Println("校验失败")
			time.Sleep(time.Second * 10)
			fmt.Println("重试")
			continue
		}

		if client.Run() {
			fmt.Println("退出")
			break
		}
		time.Sleep(time.Second * 15)
		fmt.Println("网络断开,准备重试")
		//select {
		//case <-client.connectStatus:
		//	fmt.Println("网络断开")
		//	continue
		//default:
		//
		//}
	}

}

//
//func checkLogin() {
//
//	getTokenBytes , err :=  os.ReadFile(sessionFile)
//	if err != nil {
//		fmt.Println("没有token , 准备登录")
//
//	}
//
//
//}
