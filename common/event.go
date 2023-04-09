package common

// Event 监听事件
type Event struct {
	Type int   // PUT, DELETE
	Task *Task // 任务信息
}

// NewEvent 实例化监听事件对象
func NewEvent(eType int, eTask *Task) *Event {
	return &Event{
		Type: eType,
		Task: eTask,
	}
}
