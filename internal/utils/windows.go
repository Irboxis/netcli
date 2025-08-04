//go:build windows

package utils

import (
	"nctl/interfaces"
	"nctl/internal/utils/windows"
)

// 返回 windows 平台的所有工厂函数

// 返回关于 iface 操作的工厂函数
func IfaceUtils() interfaces.Ifaces {
	return windows.Iface()
}
