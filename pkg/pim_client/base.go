package pim_client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/jroimartin/gocui"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"pim/api"
	"strings"
	"time"
)

type PimClient struct {
	clientApi api.PimServerClient
	logger    *zap.Logger
	//db        *gorm.DB
	ctx          context.Context
	currentToken string
	sessionFile  string
	rpcServerUrl string
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

	// 需要轮循事件
	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(func(gui *gocui.Gui) error {
		return layout(c, g)
	})
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
	return true
}
func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func layout(client *PimClient, g *gocui.Gui) error {

	maxX, maxY := g.Size()

	if v, err := g.SetView("Msg", 0, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Msg"
		v.Wrap = true
		v.Autoscroll = true
		v.Editable = true

	}

	return nil
}

// RunClient 只有一个参数
func RunClient(rpc string, sessionFile string) {

	// 读取session 文件

	client := &PimClient{
		ctx:          context.Background(),
		sessionFile:  sessionFile,
		rpcServerUrl: rpc,
	}

	for i := 0; i < 20; i++ {

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
			break
		}
	}

	for true {
		if client.Run() {
			fmt.Println("退出")
			break
		}
		time.Sleep(time.Second * 15)
		fmt.Println("网络断开,准备重试")
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
