//go:build linux

package utils

import (
	"nctl/interfaces"
	"nctl/internal/utils/linux"
)

// 返回 linux 平台的所有工厂函数

// 返回关于 iface 操作的工厂函数
func IfaceUtils() interfaces.Ifaces {
	return linux.Iface()
}
