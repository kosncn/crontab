package common

import "errors"

var (
	ErrorLockIsOccupied = errors.New("分布式锁已被占用")

	ErrorNoLocalIPFound = errors.New("没有找到本地网卡 IP")
)
