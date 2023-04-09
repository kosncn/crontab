package master

import (
	"context"
	"encoding/json"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"

	"crontab/common"
)

// GlobalManager 任务管理器对象
var GlobalManager = NewManager()

// Manager 任务管理器
type Manager struct {
	Client *clientV3.Client
	KV     clientV3.KV
	Lease  clientV3.Lease
}

// NewManager 实例化任务管理对象
func NewManager() *Manager {
	return &Manager{}
}

// Init 初始化任务管理对象
func (m *Manager) Init() error {
	// 实例化 etcd 配置
	config := clientV3.Config{
		Endpoints:   GlobalConfig.Endpoints,
		DialTimeout: time.Duration(GlobalConfig.DialTimeout) * time.Millisecond,
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

	return nil
}

// SaveTask 保存任务至 etcd 中
func (m *Manager) SaveTask(task *common.Task) (*common.Task, error) {
	// 序列化任务对象
	value, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	// 保存任务
	resp, err := m.KV.Put(context.TODO(), common.PathTask+task.Name, string(value), clientV3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	// 反序列化旧任务
	var oldTask *common.Task
	if resp.PrevKv != nil {
		oldTask = common.NewTask()
		_ = oldTask.Unmarshal(resp.PrevKv.Value)
	}
	return oldTask, nil
}

// DeleteTask 从 etcd 中删除任务
func (m *Manager) DeleteTask(name string) (*common.Task, error) {
	// 删除任务
	resp, err := m.KV.Delete(context.TODO(), common.PathTask+name, clientV3.WithPrevKV())
	if err != nil {
		return nil, err
	}

	// 反序列化旧任务
	var oldTask *common.Task
	if len(resp.PrevKvs) != 0 {
		oldTask = common.NewTask()
		_ = oldTask.Unmarshal(resp.PrevKvs[0].Value)
	}
	return oldTask, nil
}

// ListTask 从 etcd 中获取任务列表
func (m *Manager) ListTask() ([]*common.Task, error) {
	// 获取任务列表
	resp, err := m.KV.Get(context.TODO(), common.PathTask, clientV3.WithPrefix())
	if err != nil {
		return nil, err
	}

	// 遍历任务列表，依次反序列化
	listTask := make([]*common.Task, 0)
	for _, kv := range resp.Kvs {
		task := common.NewTask()
		if err := task.Unmarshal(kv.Value); err == nil {
			listTask = append(listTask, task)
		}
	}
	return listTask, nil
}

// KillTask 通知 worker 服务杀死任务
func (m *Manager) KillTask(name string) error {
	// 创建租约
	resp, err := m.Lease.Grant(context.TODO(), 1)
	if err != nil {
		return err
	}

	// 设置杀死任务标记
	if _, err := m.KV.Put(context.TODO(), common.PathKill+name, "", clientV3.WithLease(resp.ID)); err != nil {
		return err
	}

	return nil
}

// ListWorker 获取服务注册列表
func (m *Manager) ListWorker() ([]string, error) {
	// 初始化服务注册列表
	workerList := make([]string, 0)

	// 获取服务注册目录下所有 kv
	resp, err := m.KV.Get(context.TODO(), common.PathWorker, clientV3.WithPrefix())
	if err != nil {
		return workerList, err
	}

	// 解析所有节点的 IP
	for _, kv := range resp.Kvs {
		workerList = append(workerList, common.ExtractName(string(kv.Key), common.PathWorker))
	}

	return workerList, nil
}
