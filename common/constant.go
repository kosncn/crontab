package common

// 任务路径
const (
	// PathTask 读写任务路径
	PathTask = "/cron/task/"

	// PathKill 杀死任务路径
	PathKill = "/cron/kill/"

	// PathLock 任务分布式锁路径
	PathLock = "/cron/lock/"

	// PathWorker 服务注册路径
	PathWorker = "/cron/worker/"
)

// 响应状态
const (
	// StateSuccess 成功响应
	StateSuccess = "Success"

	// StateFailure 失败响应
	StateFailure = "Failure"
)

// 事件类型
const (
	// EventPut 保存类型
	EventPut = 0

	// EventDelete 删除类型
	EventDelete = 1

	// EventKill 杀死类型
	EventKill = 2
)
