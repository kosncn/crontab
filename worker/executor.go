package worker

import (
	"math/rand"
	"os/exec"
	"time"

	"crontab/common"
)

// GlobalExecutor 任务执行器对象
var GlobalExecutor = NewExecutor()

// Executor 任务执行器
type Executor struct{}

// NewExecutor 实例化任务执行对象
func NewExecutor() *Executor {
	return &Executor{}
}

// ExecuteTask 并发执行任务
func (e *Executor) ExecuteTask(state *common.State) {
	go func() {
		// 实例化任务执行结果对象
		result := common.NewResult()
		result.State = state

		// 创建分布式锁
		lock := GlobalManager.CreateLock(state.Task.Name)

		// 上锁前随机睡眠，保证节点间均匀竞争执行任务的机会
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		// 尝试上锁
		err := lock.TryLock()
		defer lock.UnLock()
		if err != nil { // 上锁失败
			result.StartTime = time.Now()
			result.EndTime = time.Now()
			result.Error = err
		} else { // 上锁成功
			// 记录任务开始执行时间
			result.StartTime = time.Now()

			// 执行 shell 命令
			cmd := exec.CommandContext(state.CancelCtx, GlobalConfig.BashPath, "-c", state.Task.Shell)
			output, err := cmd.Output()

			// 记录任务结束执行时间、执行结果、执行错误
			result.EndTime = time.Now()
			result.Output = output
			result.Error = err
		}

		// 推送执行执行结果到任务调度器
		GlobalScheduler.PushResult(result)
	}()
}
