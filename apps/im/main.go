package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"pim/pkg/pim_server"
	"pim/pkg/tools"
)

var versionInfo = flag.Bool("version", false, "查看版本")

var runOpt = flag.Bool("bg", false, "后台运行")
var configPath = flag.String("c", "./config.yaml", "配置文件")
var stdOutFile = flag.String("stdout", "pim.out", "stdout 重定向")

func main() {
	flag.Parse()

	if *versionInfo {

		os.Exit(0)

		return
	}

	if *runOpt {
		// 启动一个进程

		// 创建新的进程启动

		//runExec := os.Args[0]

		//fmt.Println("run exec ", runExec)

		//attr := &os.ProcAttr{Env: os.Environ(), Files: []*os.File{nil, nil, nil}}

		err := tools.Background(*stdOutFile, *configPath)

		if err != nil {
			log.Fatalln("启动失败", err)
		}

		// 启动成功
		fmt.Println("启动成功")

		return
	}

	config, err := pim_server.NewConfig(*configPath)

	if err != nil {
		log.Fatalln("配置文件错误 ", err)
	}

	pim_server.RunApp(config)

}
