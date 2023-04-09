package common

import (
	"time"

	"github.com/gorhill/cronexpr"
)

// Plan 任务调度计划
type Plan struct {
	Task     *Task                // 任务信息
	Expr     *cronexpr.Expression // 解析后的 cron 表达式
	NextTime time.Time            // 下次调度时间
}

// NewPlan 实例化任务调度计划对象
func NewPlan() *Plan {
	return &Plan{}
}

// Build 构造任务调度计划对象
func (p *Plan) Build(task *Task) error {
	// 解析 cron 表达式
	expr, err := cronexpr.Parse(task.CronExpr)
	if err != nil {
		return err
	}

	// 任务调度计划对象赋值
	p.Task = task
	p.Expr = expr
	p.NextTime = expr.Next(time.Now())

	return nil
}
