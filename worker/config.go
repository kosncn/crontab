package worker

import (
	"encoding/json"
	"os"
)

// GlobalConfig 服务配置对象
var GlobalConfig = NewConfig()

// Config 服务配置
type Config struct {
	BashPath              string   `json:"bashPath"`
	ETCDEndpoints         []string `json:"etcdEndpoints"`
	ETCDDialTimeout       int64    `json:"etcdDialTimeout"`
	MongoDBURI            string   `json:"mongoDBURI"`
	MongoDBConnectTimeout int64    `json:"mongoDBConnectTimeout"`
	ChanSize              int      `json:"chanSize"`
	BatchSize             int      `json:"batchSize"`
	LogCommitTimeout      int      `json:"logCommitTimeout"`
}

// NewConfig 实例化服务配置对象
func NewConfig() *Config {
	return &Config{}
}

// Init 初始化服务配置对象
func (c *Config) Init() error {
	// 读取配置文件
	data, err := os.ReadFile(GlobalCommand.Config)
	if err != nil {
		return err
	}

	// 反序列化至服务配置对象
	return json.Unmarshal(data, c)
}
