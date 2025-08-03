// go: build windows

package windows

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func NewWindowsNctl() (*WindowsNctl, error) {
	dll, err := windows.LoadDLL("iphlpapi.dll")
	if err != nil {
		return nil, fmt.Errorf("failed to load iphlpapi.dll: %w", err)
	}

	w := &WindowsNctl{iphlpapi: dll}

	// Load IPv4 functions
	if w.addIPAddressV4, err = dll.FindProc("AddIPAddress"); err != nil {
		return nil, err
	}
	if w.deleteIPAddressV4, err = dll.FindProc("DeleteIPAddress"); err != nil {
		return nil, err
	}
	if w.getIpAddrTable, err = dll.FindProc("GetIpAddrTable"); err != nil {
		return nil, err
	}

	// Load IPv6 functions
	if w.createUnicastIpAddressEntry, err = dll.FindProc("CreateUnicastIpAddressEntry"); err != nil {
		return nil, err
	}
	if w.deleteUnicastIpAddressEntry, err = dll.FindProc("DeleteUnicastIpAddressEntry"); err != nil {
		return nil, err
	}
	if w.getUnicastIpAddressTable, err = dll.FindProc("GetUnicastIpAddressTable"); err != nil {
		return nil, err
	}

	return w, nil
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
		return 0, fmt.Errorf("not an IPv4 address: ", ip)
	}

	return binary.LittleEndian.Uint32(ipv4), nil
}

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
