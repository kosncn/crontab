package worker

import (
	"fmt"
	"time"

	"crontab/common"
)

// GlobalScheduler 任务调度器对象
var GlobalScheduler = NewScheduler()

// Scheduler 任务调度器
type Scheduler struct {
	PlanTable  map[string]*common.Plan  // 任务调度计划表
	StateTable map[string]*common.State // 任务执行状态表
	EventChan  chan *common.Event       // 监听事件通道
	ResultChan chan *common.Result      // 任务执行结果通道
}

// NewScheduler 实例化任务调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		PlanTable:  make(map[string]*common.Plan),
		StateTable: make(map[string]*common.State),
		EventChan:  make(chan *common.Event, GlobalConfig.ChanSize),
		ResultChan: make(chan *common.Result, GlobalConfig.ChanSize),
	}
}

// Init 初始化任务调度器对象
func (s *Scheduler) Init() error {
	go s.scheduleLoop()
	return nil
}

// PushEvent 推送监听事件到任务调度器
func (s *Scheduler) PushEvent(event *common.Event) {
	s.EventChan <- event
}

// PushResult 推送执行执行结果到任务调度器
func (s *Scheduler) PushResult(result *common.Result) {
	s.ResultChan <- result
}

// handleEvent 增删改内存中维护的任务列表
func (s *Scheduler) handleEvent(event *common.Event) error {
	switch event.Type {
	case common.EventPut: // 保存任务事件
		// 实例化任务调度计划对象
		plan := common.NewPlan()
		if err := plan.Build(event.Task); err != nil {
			return err
		}
		// 保存任务调度计划
		s.PlanTable[event.Task.Name] = plan
	case common.EventDelete: // 删除任务事件
		delete(s.PlanTable, event.Task.Name)
	case common.EventKill: //杀死任务事件
		// 判断任务是否正在执行中
		if state, ok := s.StateTable[event.Task.Name]; ok {
			state.CancelFunc()
		}
	}
	return nil
}

// handleResult 处理任务执行结果
func (s *Scheduler) handleResult(result *common.Result) {
	// 删除任务执行状态
	delete(s.StateTable, result.State.Task.Name)

	// 实例化任务执行日志对象
	if result.Error != common.ErrorLockIsOccupied {
		taskLog := &common.Log{
			TaskName:  result.State.Task.Name,
			Command:   result.State.Task.Shell,
			Output:    string(result.Output),
			PlanTime:  result.State.PlanTime.UnixNano() / 1000 / 1000,
			RealTime:  result.State.RealTime.UnixNano() / 1000 / 1000,
			StartTime: result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:   result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Error != nil {
			taskLog.Error = result.Error.Error()
		} else {
			taskLog.Error = ""
		}

		// 将日志储存到 mongodb
		GlobalLogger.Save(taskLog)
	}
}

// handlePlan 处理任务调度计划
func (s *Scheduler) handlePlan(plan *common.Plan) {
	// 判断任务是否正在执行
	if _, ok := s.StateTable[plan.Task.Name]; ok {
		fmt.Println("任务正在执行:", plan.Task.Name)
		return
	}

	// 构建任务执行状态对象
	state := common.NewState()
	state.Build(plan)

	// 保存任务执行状态
	s.StateTable[state.Task.Name] = state

	fmt.Println("开始执行任务", state.Task.Name, state.PlanTime, state.RealTime)
	GlobalExecutor.ExecuteTask(state)
}

// Schedule 计算任务调度状态
func (s *Scheduler) schedule() time.Duration {
	// 判断任务调度计划表是否为空
	if len(s.PlanTable) == 0 {
		return 1 * time.Second
	}

	// 当前时间
	now := time.Now()

	// 遍历所有任务
	var nearTime *time.Time
	for _, plan := range s.PlanTable {
		// 判断任务是否需要执行
		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			// 执行任务调度计划
			s.handlePlan(plan)
			// 更新下次调度时间
			plan.NextTime = plan.Expr.Next(now)
		}
		// 统计最近需要执行的任务时间
		if nearTime == nil || nearTime.After(plan.NextTime) {
			nearTime = &plan.NextTime
		}
	}

	// 返回下次调度间隔
	return (*nearTime).Sub(now)
}

// scheduleLoop 任务调度协程
func (s *Scheduler) scheduleLoop() {
	// 实例化任务调度定时器
	timer := time.NewTimer(1 * time.Second)

	// 调度任务
	for {
		select {
		case event := <-s.EventChan: // 监听任务变化事件
			// 增删改内存中维护的任务列表
			_ = s.handleEvent(event)
		case <-timer.C: // 最近需要执行的任务到期
		case result := <-s.ResultChan: // 监听任务执行结果
			s.handleResult(result)
		}

		// 计算任务调度状态
		duration := s.schedule()

		// 重置任务调度定时器
		timer.Reset(duration)
	}
}
