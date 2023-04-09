package common

import (
	"context"
	"time"
)

// State 任务执行状态
type State struct {
	Task       *Task              // 任务信息
	PlanTime   time.Time          // 理论调度时间
	RealTime   time.Time          // 实际调度时间
	CancelCtx  context.Context    // 任务 command 的上下文
	CancelFunc context.CancelFunc // 取消任务 command 执行的函数
}

// NewState 实例化任务执行状态对象
func NewState() *State {
	return &State{}
}

// Build 构建任务执行状态对象
func (e *State) Build(plan *Plan) {
	e.Task = plan.Task
	e.PlanTime = plan.NextTime
	e.RealTime = time.Now()
	e.CancelCtx, e.CancelFunc = context.WithCancel(context.TODO())
}
