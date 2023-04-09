package common

// Log 任务执行日志
type Log struct {
	TaskName  string `json:"taskName" bson:"taskName"`   // 任务名称
	Command   string `json:"command" bson:"command"`     // 脚本命令
	Output    string `json:"output" bson:"output"`       // 执行结果
	Error     string `json:"error" bson:"error"`         // 执行错误
	PlanTime  int64  `json:"planTime" bson:"planTime"`   // 理论调度时间
	RealTime  int64  `json:"realTime" bson:"realTime"`   // 实际调度时间
	StartTime int64  `json:"startTime" bson:"startTime"` // 开始执行时间
	EndTime   int64  `json:"endTime" bson:"endTime"`     // 结束执行时间
}

// NewLog 实例化任务执行日志对象
func NewLog() *Log {
	return &Log{}
}

// LogFilter 任务执行日志过滤条件
type LogFilter struct {
	TaskName string `bson:"taskName"`
}

// NewLogFilter 实例化任务执行日志过滤条件对象
func NewLogFilter(name string) *LogFilter {
	return &LogFilter{TaskName: name}
}

// LogSorter 任务执行日志排序规则
type LogSorter struct {
	StartTime int64 `bson:"startTime"` // 倒序: {startTime: -1}
}

// NewLogSorter 实例化任务执行日志排序规则对象
func NewLogSorter(startTime int64) *LogSorter {
	return &LogSorter{StartTime: startTime}
}
