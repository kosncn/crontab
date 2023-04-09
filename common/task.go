package common

import (
	"encoding/json"
	"strings"
)

// Task 任务
type Task struct {
	Name     string `json:"name"`     // 任务名称
	Shell    string `json:"shell"`    // shell 命令
	CronExpr string `json:"cronExpr"` // cron 表达式
}

// NewTask 实例化任务对象
func NewTask() *Task {
	return &Task{}
}

// Unmarshal 反序列化任务数据
func (t *Task) Unmarshal(data []byte) error {
	err := json.Unmarshal(data, t)
	return err
}

// ExtractName 从 etcd 的 key 中提取任务名称
func ExtractName(key string, path string) string {
	return strings.TrimPrefix(key, path)
}
