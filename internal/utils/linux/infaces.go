//go:build linux

package linux

import (
	"bytes"
	"fmt"
	"nctl/internal/utils"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/godbus/dbus"
	"github.com/vishvananda/netlink"
)

// 检查接口的存在性
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

// 增加ip
func (u *UnixNctl) AddIP(iface string, ipnet *net.IPNet) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("failed to get interfaces: %v", err)
	}

	addr := &netlink.Addr{IPNet: ipnet}

	return netlink.AddrAdd(link, addr)
}

// 删除已存在的ip
func (u *UnixNctl) DelIP(iface string, ipnet *net.IPNet) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	addr := &netlink.Addr{IPNet: ipnet}

	return netlink.AddrDel(link, addr)
}

// 设置默认网关
func (u *UnixNctl) SetGateway(iface string, gateway net.IP) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("failed to get interface '%s': %w", iface, err)
	}

	routes, err := netlink.RouteList(nil, netlink.FAMILY_ALL)
	if err != nil {
		return fmt.Errorf("failed to list routes: %w", err)
	}
	for _, r := range routes {
		if r.Gw != nil && r.Dst == nil {
			if err := netlink.RouteDel(&r); err != nil {
				if !strings.Contains(err.Error(), "no such process") {
					fmt.Fprintf(os.Stderr, "Warning: failed to delete old default gateway: %v\n", err)
				}
			}
		}
	}

	newRoute := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Gw:        gateway,
		Dst:       nil,
		Scope:     netlink.SCOPE_UNIVERSE,
		Protocol:  syscall.RTPROT_STATIC,
	}
	if err := netlink.RouteAdd(newRoute); err != nil {
		return fmt.Errorf("failed to set gateway: %w", err)
	}

	return nil
}

// 增加dns
func (u *UnixNctl) AddDNS(iface string, dnsIP net.IP) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("failed to connect to D-Bus system bus: %w", err)
	}
	defer conn.Close()

	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("failed to get interface '%s': %w", iface, err)
	}

	obj := conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1")
	var currentDNS [][]byte
	err = obj.Call("org.freedesktop.resolve1.Manager.GetLinkDNS", 0, link.Attrs().Index).Store(&currentDNS)
	if err != nil {
		return fmt.Errorf("failed to get current DNS servers: %w", err)
	}

	isExist := false
	for _, ipBytes := range currentDNS {
		if bytes.Equal(ipBytes, dnsIP) {
			isExist = true
			break
		}
	}

	if isExist {
		fmt.Fprintf(os.Stdout, "DNS server '%s' is already set for interface '%s'\n", dnsIP.String(), iface)
		return nil
	}

	newDNSes := make([][]byte, 0, len(currentDNS)+1)
	for _, ipBytes := range currentDNS {
		newDNSes = append(newDNSes, ipBytes)
	}
	newDNSes = append(newDNSes, dnsIP)

	call := obj.Call("org.freedesktop.resolve1.Manager.SetLinkDNS", 0, link.Attrs().Index, newDNSes)
	if call.Err != nil {
		return fmt.Errorf("failed to set DNS server '%s': %w", dnsIP.String(), call.Err)
	}

	fmt.Fprintf(os.Stderr, "Successfully added DNS server '%s' for interface '%s'\n", dnsIP.String(), iface)
	return nil
}

// 删除指定dns
func (u *UnixNctl) DelDNS(iface string, dnsIP net.IP) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return fmt.Errorf("failed to connect to D-Bus system bus: %w", err)
	}
	defer conn.Close()

	link, err := netlink.LinkByName(iface)
	if err != nil {
		return fmt.Errorf("failed to get interface '%s': %w", iface, err)
	}

	obj := conn.Object("org.freedesktop.resolve1", "/org/freedesktop/resolve1")
	var currentDNS [][]byte
	err = obj.Call("org.freedesktop.resolve1.Manager.GetLinkDNS", 0, link.Attrs().Index).Store(&currentDNS)
	if err != nil {
		return fmt.Errorf("failed to get current DNS servers: %w", err)
	}

	found := false
	newDNSes := make([][]byte, 0, len(currentDNS))
	for _, ipBytes := range currentDNS {
		if !bytes.Equal(ipBytes, dnsIP) {
			newDNSes = append(newDNSes, ipBytes)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("DNS server '%s' not found for interface '%s'", dnsIP.String(), iface)
	}

	call := obj.Call("org.freedesktop.resolve1.Manager.SetLinkDNS", 0, link.Attrs().Index, newDNSes)
	if call.Err != nil {
		return fmt.Errorf("failed to delete DNS server '%s': %w", dnsIP.String(), call.Err)
	}

	fmt.Printf("Successfully deleted DNS server '%s' for interface '%s'\n", dnsIP.String(), iface)
	return nil
}

// 编译时接口检查
var _ utils.Ifaces = (*UnixNctl)(nil)
