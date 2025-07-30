package utils

import (
	"net"

	"github.com/jackpal/gateway"
)

// 传入接口的mac地址，获取其默认网关地址
func GetGW(iface *net.Interface) (ipv4Gateway string, ipv6Gateway string) {
	ipv4Gateway = "N/A"
	ipv6Gateway = "N/A"

	// 获取所有系统默认网关
	systemGateways, err := gateway.DiscoverGateways()
	if err != nil {
		return ipv4Gateway, ipv6Gateway
	}

	// 获取目标接口的所有 IP 地址
	addrs, err := iface.Addrs()
	if err != nil {
		// 记录警告但继续，因为可能只是当前接口没有地址
		return ipv4Gateway, ipv6Gateway
	}

	// 尝试通过子网将网关与目标接口关联
	for _, gwIP := range systemGateways {
		isIPv4 := gwIP.To4() != nil
		isIPv6 := gwIP.To16() != nil && gwIP.To4() == nil // 纯 IPv6

		if !isIPv4 && !isIPv6 {
			continue // 跳过未知 IP 类型
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue // 不是 IPNet 地址
			}

			if isIPv4 && ipNet.IP.To4() != nil {
				if ipv4Gateway == "N/A" && ipNet.Contains(gwIP) { // 如果 IPv4 网关还没找到且子网匹配
					ipv4Gateway = gwIP.String()
					break
				}
			} else if isIPv6 && ipNet.IP.To16() != nil && ipNet.IP.To4() == nil {
				if ipv6Gateway == "N/A" && ipNet.Contains(gwIP) { // 如果 IPv6 网关还没找到且子网匹配
					ipv6Gateway = gwIP.String()
					break
				}
			}
		}
		// 如果 IPv4 和 IPv6 网关都已找到，就可以提前返回
		if ipv4Gateway != "N/A" && ipv6Gateway != "N/A" {
			break
		}
	}

	return ipv4Gateway, ipv6Gateway
}
