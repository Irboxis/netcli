//go:build linux || darwin

package utils

import (
	"fmt"
	"net"
)

// 工厂函数
func newNctltools() Ifaces {
	return &UnixNctl{}
}

func (u *UnixNctl) IsExistingIface(iface string) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %v", err)
	}
	for _, i := range ifaces {
		if i.Name == iface {
			return nil
		}
	}
	return fmt.Errorf("interface %s not found", iface)
}

// 编译时接口检查
var _ Ifaces = (*UnixNctl)(nil)
