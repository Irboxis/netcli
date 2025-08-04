//go:build linux

package linux

import (
	"nctl/interfaces"
)

type UnixNctl struct{}

// 工厂函数
func Iface() interfaces.Ifaces {
	return &UnixNctl{}
}
