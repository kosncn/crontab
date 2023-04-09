package worker

import (
	"context"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientV3 "go.etcd.io/etcd/client/v3"

	"crontab/common"
)

// GlobalManager 任务管理器对象
var GlobalManager = NewManager()

// Manager 任务管理器
type Manager struct {
	Client  *clientV3.Client
	KV      clientV3.KV
	Lease   clientV3.Lease
	Watcher clientV3.Watcher
}

// NewManager 实例化任务管理对象
func NewManager() *Manager {
	return &Manager{}
}

// Init 初始化任务管理对象
func (m *Manager) Init() error {
	// 实例化 etcd 配置
	config := clientV3.Config{
		Endpoints:   GlobalConfig.ETCDEndpoints,
		DialTimeout: time.Duration(GlobalConfig.ETCDDialTimeout) * time.Millisecond,
	}

	// 创建 etcd 客户端
	client, err := clientV3.New(config)
	if err != nil {
		return err
	}

	// 任务管理器对象赋值
	m.Client = client
	m.KV = clientV3.NewKV(client)
	m.Lease = clientV3.NewLease(client)
	m.Watcher = clientV3.NewWatcher(client)

	// 监听 etcd 中增删任务变化
	if err := m.WatchTask(); err != nil {
		return err
	}

	// 监听 etcd 中杀死任务变化事件
	go m.WatchKill()

	return nil
}

// WatchTask 监听 etcd 中增删任务变化
func (m *Manager) WatchTask() error {
	// 获取定时任务列表
	resp, err := m.KV.Get(context.TODO(), common.PathTask, clientV3.WithPrefix())
	if err != nil {
		return err
	}

	// 遍历任务列表，依次反序列化
	for _, kv := range resp.Kvs {
		task := common.NewTask()
		if err := task.Unmarshal(kv.Value); err != nil {
			continue
		}
		event := common.NewEvent(common.EventPut, task)

		// 推送监听事件到任务调度器
		GlobalScheduler.PushEvent(event)
	}

	// 监听 etcd 中任务变化事件
	go m.watchEvent(resp.Header.Revision)

	return nil
}

// WatchKill 监听 etcd 中杀死任务变化
func (m *Manager) WatchKill() {
	// 监听任务变化事件
	watchChan := m.Watcher.Watch(context.TODO(), common.PathKill, clientV3.WithPrefix())

	// 处理监听事件
	for resp := range watchChan {

		// 遍历监听事件列表，依次反序列化
		for _, e := range resp.Events {
			switch e.Type {
			case mvccpb.PUT: // 杀死任务事件
				task := common.NewTask()
				task.Name = common.ExtractName(string(e.Kv.Key), common.PathKill)
				event := common.NewEvent(common.EventKill, task)
				// 推送监听事件到任务调度器
				GlobalScheduler.PushEvent(event)
			case mvccpb.DELETE: // kill 标记过期，被自动删除

			}
		}
	}
}

// watchEvent 监听 etcd 中任务变化事件
func (m *Manager) watchEvent(revision int64) {
	// 监听任务变化事件
	watchChan := m.Watcher.Watch(context.TODO(), common.PathTask, clientV3.WithPrefix(), clientV3.WithRev(revision+1))

	// 处理监听事件
	for resp := range watchChan {

		// 遍历监听事件列表，依次反序列化
		for _, e := range resp.Events {
			var event *common.Event
			task := common.NewTask()

			switch e.Type {
			case mvccpb.PUT: // 保存任务事件
				if err := task.Unmarshal(e.Kv.Value); err != nil {
					continue
				}
				event = common.NewEvent(common.EventPut, task)
			case mvccpb.DELETE: // 删除任务事件
				task.Name = common.ExtractName(string(e.Kv.Key), common.PathTask)
				event = common.NewEvent(common.EventDelete, task)
			}

			// 推送监听事件到任务调度器
			GlobalScheduler.PushEvent(event)
		}
	}
}

// CreateLock 创建分布式锁
func (m *Manager) CreateLock(taskName string) *Lock {
	return NewLock(taskName, m.KV, m.Lease)
}
