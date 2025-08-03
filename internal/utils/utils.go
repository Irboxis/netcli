package utils

import (
	"net"
)

type Ifaces interface {
	// 检查接口存在性
	IsExistingIface(iface string) error
	// ip 的增删
	AddIP(iface string, ipnet *net.IPNet) error
	DelIP(iface string, ipnet *net.IPNet) error
	// dns 的增删
	AddDNS(iface string, dnsIP net.IP) error
	DelDNS(iface string, dnsIP net.IP) error
	// 网关设置操作
	SetGateway(iface string, gateway net.IP) error
}

// 统一返回各个文件的工厂函数
// 所有需要实现 Ifaces 接口和 xxxNctl 结构体的文件，都必须添加 newNctltools 函数
func NewNctlUtils() Ifaces {
	return NewNctlUtils()
}
