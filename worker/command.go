package worker

import (
	"flag"
)

// GlobalCommand 命令行参数对象
var GlobalCommand = NewCommand()

// Command 命令行参数
type Command struct {
	Config string
}

// NewCommand 实例化命令行参数对象
func NewCommand() *Command {
	return &Command{}
}

// Init 初始化命令行参数对象
func (c *Command) Init() error {
	// 解析命令行参数
	// worker.exe -config="./master.json"
	config := flag.String("config", "./config/worker.json", "输入服务配置文件路径")
	flag.Parse()

	// 命令行参数对象赋值
	c.Config = *config

	return nil
}
