package interfaces

import (
	"net"
)

type Ifaces interface {
	// 检查接口存在性
	IsExistingIface(iface string) error
	// ip 的增删
	AddIP(iface string, ipnet *net.IPNet) error
	DelIP(iface string, ipnet *net.IPNet) error
	// 覆盖接口的 ip 设置
	SetIPs(ifaceName string, ipnets []*net.IPNet) error

	// dns 的增删
	AddDNS(iface string, dnsIP net.IP) error
	DelDNS(iface string, dnsIP net.IP) error
	// 覆盖接口的dns设置
	SetDNSs(ifaceName string, dnsIPs []net.IP) error

	// 网关设置操作
	SetGateway(iface string, gateway net.IP) error
}
