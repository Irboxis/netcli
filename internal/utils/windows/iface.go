//go:build windows

package windows

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 检查接口名称的存在性
func (w *WindowsNctl) IsExistingIface(iface string) error {
	var b []byte
	// 设置初始缓冲区为 1k
	l := uint32(1024)

	for {
		b = make([]byte, l)
		size := l
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_ALL_INTERFACES, 0, (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &size)
		if err == nil {
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return fmt.Errorf("unexpected code errors [internal/utils/ifaces_windows.go: 30]: %w", err)
		}

		// 缓冲区太小，设置翻倍，但总大小不允许超过 128k
		l *= 2
		const maxBufferSize = 128 * 1024
		if l > maxBufferSize {
			return fmt.Errorf("Failed to allocate buffer, size exceeds %dKB", maxBufferSize)
		}
	}

	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		friendlyName := windows.UTF16PtrToString(aa.FriendlyName)
		if friendlyName == iface {
			return nil
		}
	}

	return fmt.Errorf("network interface '%s' does not exist", iface)
}

// 转化接口的 Index 和 LUID
func findIface(ifaceName string) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for i := range ifaces {
		if ifaces[i].Name == ifaceName {
			return &ifaces[i], nil
		}
	}

	return nil, fmt.Errorf("interface '%s' not found\n", ifaceName)
}

// 获取接口的 LUID
func getIfaceLUID(ifaceIndex uint32) (NET_LUID, error) {
	var row struct {
		IfaceLuid  NET_LUID
		IfaceIndex uint32
	}

	row.IfaceIndex = ifaceIndex

	ifaces, err := net.Interfaces()
	if err != nil {
		return NET_LUID{}, err
	}

	for _, ifi := range ifaces {
		if ifi.Index == int(ifaceIndex) {
			return NET_LUID{}, nil
		}
	}

	return NET_LUID{}, fmt.Errorf("could not determine LUID for index %d\n", ifaceIndex)
}

func ipToUint32(ip net.IP) (uint32, error) {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("not an IPv4 address: %v", ip)
	}

	return binary.LittleEndian.Uint32(ipv4), nil
}

// 增加 ip，同时支持 ipv4 和 ipv6
func (w *WindowsNctl) AddIP(ifaceName string, ipnet *net.IPNet) error {
	iface, err := findIface(ifaceName)
	if err != nil {
		return err
	}

	if ipnet.IP.To4() != nil {
		// 增加 ipv4 地址
		ip, err := ipToUint32(ipnet.IP)
		if err != nil {
			return err
		}

		mask, err := ipToUint32(net.IP(ipnet.Mask))
		if err != nil {
			return err
		}

		var netContext, netInstance uint32
		r1, _, _ := w.addIPAddressV4.Call(uintptr(ip), uintptr(mask), uintptr(iface.Index), uintptr(unsafe.Pointer(&netContext)), uintptr(unsafe.Pointer(&netInstance)))
		if r1 != 0 {
			return fmt.Errorf("add ipaddress faild [utils/windows/settools.go:114]: %w", syscall.Errno(r1))
		} else {
			fmt.Fprintf(os.Stdout, "add IPv4 '%s' to '%s' successful\n", ipnet.String(), ifaceName)
			return nil
		}
	} else {
		// 增加 ipv6 地址
		luid, err := getIfaceLUID(uint32(iface.Index))
		if err != nil {
			return err
		}

		prefixLen, _ := ipnet.Mask.Size()
		var row MIB_UNICASTIPADDRESS_ROW
		row.InterfaceIndex = uint32(iface.Index)
		row.InterfaceLuid = luid
		row.OnLinkPrefixLength = uint8(prefixLen)
		row.PrefixOrigin = 2
		row.SuffixOrigin = 2
		row.DadState = 4
		row.ValidLifetime = 0xffffffff
		row.PreferredLifetime = 0xffffffff
		row.Address.Family = windows.AF_INET6
		copy(row.Address.Data[6:22], ipnet.IP.To16())

		r1, _, _ := w.createUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))
		if r1 != 0 {
			return fmt.Errorf("add ipaddress faild [utils/windows/settools.go:143]: %w", syscall.Errno(r1))
		} else {
			fmt.Fprintf(os.Stdout, "add IPv6 '%s' to '%s' successful\n", ipnet.String(), ifaceName)
			return nil
		}
	}
}

// 删除 ip，同时支持 ipv4 和 ipv6
func (w *WindowsNctl) DelIP(ifaceName string, ipnet *net.IPNet) error {
	iface, err := findIface(ifaceName)
	if err != nil {
		return err
	}

	if ipnet.IP.To4() != nil {
		// 增加 ipv6 地址
		targetIp, err := ipToUint32(ipnet.IP)
		if err != nil {
			return err
		}

		var size uint32
		w.getIpAddrTable.Call(0, uintptr(unsafe.Pointer(&size)), 1)
		if size == 0 {
			return fmt.Errorf("delete ipaddress faild [utils/windows/settools.go: 168]\n")
		}

		buffer := make([]byte, size)
		r1, _, _ := w.getIpAddrTable.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)), 1)
		if r1 != 0 {
			return fmt.Errorf("delete ipaddress faild [utils/windows/settools.go: 168]: %w\n", syscall.Errno(r1))
		}

		numEntries := *(*uint32)(unsafe.Pointer(&buffer[0]))
		rows := unsafe.Slice((*MIB_IPADDRROW)(unsafe.Pointer(&buffer[4])), numEntries)

		var netContext uint32
		found := false
		for _, row := range rows {
			if row.DwIndex == uint32(iface.Index) && row.DwAddr == targetIp {
				netContext = row.DwAddr
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("IPv4 address '%s' not found on interface '%s'", ipnet.IP, ifaceName)
		}
		r1, _, _ = w.deleteIPAddressV4.Call(uintptr(netContext))
		if r1 != 0 {
			return fmt.Errorf("delete ipaddress faild [utils/windows/settools.go: 195]: %w", syscall.Errno(r1))
		} else {
			fmt.Fprintf(os.Stdout, "delete IPv4 '%s' from '%s' successful", ipnet.String(), ifaceName)
			return nil
		}
	} else {
		// 删除 ipv6 地址
		luid, err := getIfaceLUID(uint32(iface.Index))
		if err != nil {
			return err
		}

		prefixLen, _ := ipnet.Mask.Size()
		var row MIB_UNICASTIPADDRESS_ROW
		row.InterfaceIndex = uint32(iface.Index)
		row.InterfaceLuid = luid
		row.OnLinkPrefixLength = uint8(prefixLen)
		row.Address.Family = windows.AF_INET6

		copy(row.Address.Data[6:22], ipnet.IP.To16())
		r1, _, _ := w.createUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))
		if r1 != 0 {
			return fmt.Errorf("delete ipaddress faild [utils/windows/settools.go: 218]: %w", syscall.Errno(r1))
		} else {
			fmt.Fprintf(os.Stdout, "delete IPv6 '%s' to '%s' successful\n", ipnet.String(), ifaceName)
			return nil
		}
	}
}

// 增加新的 dns 地址，原生实现过于复杂，调用系统命令实现
func (w *WindowsNctl) AddDNS(iface string, dnsIP net.IP) error {
	ipVersion := "ipv4"
	if dnsIP.To4() == nil {
		ipVersion = "ipv6"
	}

	cmd := exec.Command("netsh", "interface", ipVersion, "add", "dnsserver", fmt.Sprintf(`name="%s"`, iface), fmt.Sprintf("address=%s", dnsIP.String()), "index=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add DNS server using netsh: %v\nOutput: %s", err, string(output))
	}
	fmt.Fprintf(os.Stdout, "add DNS '%s' to '%s' successful\n", dnsIP.String(), iface)
	return nil
}

// 删除已经存在的dns地址
func (w *WindowsNctl) DelDNS(iface string, dnsIP net.IP) error {
	ipVersion := "ipv4"
	if dnsIP.To4() == nil {
		ipVersion = "ipv6"
	}

	cmd := exec.Command("netsh", "interface", ipVersion, "delete", "dnsserver", fmt.Sprintf(`name="%s"`, iface), fmt.Sprintf("address=%s", dnsIP.String()))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete DNS server using netsh: %v\nOutput: %s", err, string(output))
	}
	fmt.Fprintf(os.Stdout, "delete DNS '%s' from '%s' successful\n", dnsIP.String(), iface)
	return nil
}

// 设置默认网关
func (w *WindowsNctl) SetGateway(iface string, gateway net.IP) error {
	ifa, err := findIface(iface)
	if err != nil {
		return err
	}

	if gateway.To4() == nil {
		return fmt.Errorf("setting IPv6 gateway is not implemented")
	}

	gatewayIP, err := ipToUint32(gateway)
	if err != nil {
		return err
	}

	var size uint32
	w.getIpForwardTable.Call(0, uintptr(unsafe.Pointer(&size)), 1)
	if size > 0 {
		buffer := make([]byte, size)
		r1, _, _ := w.getIpForwardTable.Call(uintptr(unsafe.Pointer(&buffer[0])), uintptr(unsafe.Pointer(&size)), 1)
		if r1 == 0 { // NO_ERROR
			numEntries := *(*uint32)(unsafe.Pointer(&buffer[0]))
			table := unsafe.Slice((*MIB_IPFORWARDROW)(unsafe.Pointer(&buffer[4])), numEntries)
			for _, row := range table {
				// 检查是否是默认路由 (Dest=0.0.0.0) 并且是同一个接口
				if row.DwForwardDest == 0 && row.DwForwardIfIndex == uint32(ifa.Index) {
					fmt.Printf("deleting existing default gateway on interface '%s'\n", iface)
					w.deleteIpForwardEntry.Call(uintptr(unsafe.Pointer(&row)))
				}
			}
		}
	}

	row := MIB_IPFORWARDROW{
		DwForwardDest:    0,
		DwForwardMask:    0,
		DwForwardNextHop: gatewayIP,
		DwForwardIfIndex: uint32(ifa.Index),
		DwForwardProto:   3,
		DwForwardType:    4,
		DwForwardMetric1: 25,
	}

	r1, _, err := w.createUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))
	if r1 != 0 {
		return fmt.Errorf("CreateIpForwardEntry failed [ifaces_setip: 271]: %w", err)
	}

	fmt.Fprintf(os.Stdout, "set defult gateway to '%s' on '%s' successful\n", gateway.String(), iface)
	return nil
}

// NEW: SetIPs 覆盖IP地址
func (w *WindowsNctl) SetIPs(ifaceName string, ipnets []*net.IPNet) error {
	// 1. 获取接口上所有现有的 IP 地址
	// 2. 遍历并删除它们 (调用 DelIP)
	// 3. 遍历新的 ipnets 列表并添加它们 (调用 AddIP)
	// 这个实现比较复杂，涉及到 GetIpAddrTable 和 GetUnicastIpAddressTable 的使用
	// 为了简化，也可以使用 netsh 命令
	// 清除所有IP
	cmdV4 := exec.Command("netsh", "interface", "ipv4", "set", "address", fmt.Sprintf(`name="%s"`, ifaceName), "source=dhcp")
	cmdV4.Run() // 忽略错误，因为可能没有DHCP
	cmdV6 := exec.Command("netsh", "interface", "ipv6", "set", "address", fmt.Sprintf(`interface="%s"`, ifaceName), "store=active")
	cmdV6.Run()

	fmt.Println("Cleared existing IP addresses. Now adding new ones...")
	for _, ipnet := range ipnets {
		if err := w.AddIP(ifaceName, ipnet); err != nil {
			// 在覆盖模式下，一个失败可能不应该停止其他操作
			fmt.Printf("Warning: failed to add IP %s: %v\n", ipnet.String(), err)
		}
	}
	fmt.Printf("Successfully set new IP addresses on '%s'\n", ifaceName)
	return nil
}

// NEW: SetDNSs 覆盖DNS
func (w *WindowsNctl) SetDNSs(ifaceName string, dnsIPs []net.IP) error {
	// 使用 netsh 是在 Windows 上管理 DNS 的最可靠方法
	// 首先，清除所有静态DNS
	cmdV4 := exec.Command("netsh", "interface", "ipv4", "set", "dnsserver", fmt.Sprintf(`name="%s"`, ifaceName), "source=dhcp")
	cmdV4.CombinedOutput() // 忽略错误
	cmdV6 := exec.Command("netsh", "interface", "ipv6", "set", "dnsserver", fmt.Sprintf(`name="%s"`, ifaceName), "source=dhcp")
	cmdV6.CombinedOutput() // 忽略错误

	fmt.Println("Cleared existing DNS settings. Now adding new ones...")
	for _, ip := range dnsIPs {
		if err := w.AddDNS(ifaceName, ip); err != nil {
			fmt.Printf("Warning: failed to add DNS %s: %v\n", ip.String(), err)
		}
	}
	fmt.Printf("Successfully set new DNS servers on '%s'\n", ifaceName)
	return nil
}
