package worker

import (
	"context"
	"net"
	"time"

	clientV3 "go.etcd.io/etcd/client/v3"

	"crontab/common"
)

// GlobalRegister 服务注册对象
var GlobalRegister = NewRegister()

// Register 服务注册
type Register struct {
	Client  *clientV3.Client
	KV      clientV3.KV
	Lease   clientV3.Lease
	LocalIP string
}

// NewRegister 实例化服务注册对象
func NewRegister() *Register {
	return &Register{}
}

// Init 初始化服务注册对象
func (r *Register) Init() error {
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

	// 获取本机 IPV4 地址
	localIP, err := r.GetLocalIP()
	if err != nil {
		return err
	}

	// 服务注册对象赋值
	r.Client = client
	r.KV = clientV3.NewKV(client)
	r.Lease = clientV3.NewLease(client)
	r.LocalIP = localIP

	// 注册服务并自动续租
	go r.KeepOnline()

	return nil
}

// GetLocalIP 获取本机 IPV4 地址
func (r *Register) GetLocalIP() (string, error) {
	// 获取所有网卡信息
	address, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	// 获取第一个非环回地址的 ipv4 地址
	for _, addr := range address {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", common.ErrorNoLocalIPFound
}

// KeepOnline 注册服务并自动续租
func (r *Register) KeepOnline() {
	// 回滚函数
	rollback := func(cancelFunc context.CancelFunc) {
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}
	}

	for {
		// 创建租约
		grantResp, err := r.Lease.Grant(context.TODO(), 10)
		if err != nil {
			rollback(nil)
			continue
		}

		// 自动续租
		keepAliveChan, err := r.Lease.KeepAlive(context.TODO(), grantResp.ID)
		if err != nil {
			rollback(nil)
			continue
		}

		// 创建上下文
		ctx, cancel := context.WithCancel(context.TODO())

		// 将本机 IP 注册到 etcd
		if _, err := r.KV.Put(ctx, common.PathWorker+r.LocalIP, "", clientV3.WithLease(grantResp.ID)); err != nil {
			rollback(cancel)
		}

		// 处理自动续租应答
		for keepAliveResp := range keepAliveChan {
			if keepAliveResp == nil {
				break
			}
		}

		rollback(cancel)
	}
}
