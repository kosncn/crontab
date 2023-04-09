package worker

import (
	"context"

	clientV3 "go.etcd.io/etcd/client/v3"

	"crontab/common"
)

// Lock 分布式锁
type Lock struct {
	TaskName string
	KV       clientV3.KV
	Lease    clientV3.Lease
	LeaseID  clientV3.LeaseID
	Cancel   context.CancelFunc // 用于终止自动续租
	isLocked bool               // 是否上锁成功
}

// NewLock 实例化分布式锁对象
func NewLock(taskName string, kv clientV3.KV, lease clientV3.Lease) *Lock {
	return &Lock{
		TaskName: taskName,
		KV:       kv,
		Lease:    lease,
	}
}

// TryLock 尝试上分布式锁
func (l *Lock) TryLock() error {
	// 创建租约（5秒）
	grantResp, err := l.Lease.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}

	// 创建上下文对象，用于取消自动续租
	ctx, cancel := context.WithCancel(context.TODO())

	// 启动自动续租
	keepRespChan, err := l.Lease.KeepAlive(ctx, grantResp.ID)
	if err != nil {
		cancel()                                            // 取消自动续租
		_, _ = l.Lease.Revoke(context.TODO(), grantResp.ID) // 释放租约
		return err
	}

	// 处理自动续租应答
	go func() {
		for keepResp := range keepRespChan {
			if keepResp == nil {
				break
			}
		}
	}()

	// 创建 txn 事务
	txn := l.KV.Txn(context.TODO())

	// 事务抢分布式锁
	key := common.PathLock + l.TaskName
	txn.If(clientV3.Compare(clientV3.CreateRevision(key), "=", 0)).
		Then(clientV3.OpPut(key, "", clientV3.WithLease(grantResp.ID))).
		Else(clientV3.OpGet(key))

	// 提交事务
	txnResp, err := txn.Commit()
	if err != nil {
		cancel()                                            // 取消自动续租
		_, _ = l.Lease.Revoke(context.TODO(), grantResp.ID) // 释放租约
		return err
	}

	// 分布式锁已被占用
	if !txnResp.Succeeded {
		cancel()                                            // 取消自动续租
		_, _ = l.Lease.Revoke(context.TODO(), grantResp.ID) // 释放租约
		return common.ErrorLockIsOccupied

	}

	// 抢分布式锁成功
	l.LeaseID = grantResp.ID
	l.Cancel = cancel
	l.isLocked = true
	return nil
}

// UnLock 释放分布式锁
func (l *Lock) UnLock() {
	if l.isLocked {
		l.Cancel()                                       // 取消自动续租
		_, _ = l.Lease.Revoke(context.TODO(), l.LeaseID) // 释放租约
	}
}
