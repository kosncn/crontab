package main

import (
	"log"

	"crontab/worker"
)

// 初始化并启动 worker 服务
func main() {
	// 初始化命令行参数
	if err := worker.GlobalCommand.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化服务配置
	if err := worker.GlobalConfig.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化服务注册器
	if err := worker.GlobalRegister.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化日志管理器
	if err := worker.GlobalLogger.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化任务调度器
	if err := worker.GlobalScheduler.Init(); err != nil {
		log.Fatalln(err)
	}

	// 初始化任务管理器
	if err := worker.GlobalManager.Init(); err != nil {
		log.Fatalln(err)
	}

	select {}
}
