package common

import (
	"time"
)

// Result 任务执行结果
type Result struct {
	State     *State    // 任务信息
	Output    []byte    // 执行结果
	Error     error     // 执行错误
	StartTime time.Time // 开始执行时间
	EndTime   time.Time // 结束执行时间
}

// NewResult 实例化任务执行结果对象
func NewResult() *Result {
	return &Result{}
}
