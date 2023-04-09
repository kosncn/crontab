package main

import (
	"log"

	"crontab/master"
)

// 初始化并启动 master 服务
func main() {
	// 初始化命令行参数
	if err := master.GlobalCommand.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化服务配置
	if err := master.GlobalConfig.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化日志管理器
	if err := master.GlobalLogger.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化任务管理器
	if err := master.GlobalManager.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化服务
	if err := master.GlobalServer.Init(); err != nil {
		log.Fatalln(err)
	}

	// 启动服务
	if err := master.GlobalServer.HTTPServer.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
