// go: build windows

package windows

import (
	"fmt"
	"nctl/interfaces"
	"os"

	"golang.org/x/sys/windows"
)

// 编译时接口检查
var _ interfaces.Ifaces = (*WindowsNctl)(nil)

// 工厂函数
func Iface() interfaces.Ifaces {
	w, err := newWindowsNctl()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize Windows network controller: %v", err)
	}
	return w
}

type WindowsNctl struct {
	iphlpapi *windows.DLL
	// 针对 ipv4
	addIPAddressV4    *windows.Proc
	deleteIPAddressV4 *windows.Proc
	getIpAddrTable    *windows.Proc

	// 针对 ipv6
	createUnicastIpAddressEntry *windows.Proc
	deleteUnicastIpAddressEntry *windows.Proc
	getUnicastIpAddressTable    *windows.Proc

	// 针对网关
	getIpForwardTable    *windows.Proc
	createIpForwardEntry *windows.Proc
	deleteIpForwardEntry *windows.Proc

	//convertInterfaceIndexToLuid *windows.Proc
}

// MIB_IPADDRROW for IPv4
type MIB_IPADDRROW struct {
	DwAddr, DwIndex, DwMask, DwBCastAddr, DwReasmSize uint32
	Unused1, WType                                    uint16
}

// NET_LUID for IPv6
type NET_LUID struct {
	Value uint64
}

// SOCKADDR_INET for IPv6 address structure
type SOCKADDR_INET struct {
	Family uint16   // AF_INET or AF_INET6
	Data   [26]byte // Enough space for IPv6 sockaddr
}

// MIB_UNICASTIPADDRESS_ROW for IPv6
type MIB_UNICASTIPADDRESS_ROW struct {
	Address            SOCKADDR_INET
	InterfaceLuid      NET_LUID
	InterfaceIndex     uint32
	PrefixOrigin       uint32
	SuffixOrigin       uint32
	ValidLifetime      uint32
	PreferredLifetime  uint32
	OnLinkPrefixLength uint8
	SkipAsSource       bool
	DadState           uint32
	ScopeId            uint32
	CreationTimeStamp  int64
}

// 网关操作的结构体
type MIB_IPFORWARDROW struct {
	DwForwardDest      uint32
	DwForwardMask      uint32
	DwForwardPolicy    uint32
	DwForwardNextHop   uint32
	DwForwardIfIndex   uint32
	DwForwardType      uint32
	DwForwardProto     uint32
	DwForwardAge       uint32
	DwForwardNextHopAS uint32
	DwForwardMetric1   uint32
	DwForwardMetric2   uint32
	DwForwardMetric3   uint32
	DwForwardMetric4   uint32
	DwForwardMetric5   uint32
}

func newWindowsNctl() (*WindowsNctl, error) {
	dll, err := windows.LoadDLL("iphlpapi.dll")
	if err != nil {
		return nil, fmt.Errorf("failed to load iphlpapi.dll: %w", err)
	}
	w := &WindowsNctl{iphlpapi: dll}

	// 简化函数查找逻辑
	procs := []struct {
		name string
		ptr  **windows.Proc
	}{
		{"AddIPAddress", &w.addIPAddressV4},
		{"DeleteIPAddress", &w.deleteIPAddressV4},
		{"GetIpAddrTable", &w.getIpAddrTable},
		{"CreateUnicastIpAddressEntry", &w.createUnicastIpAddressEntry},
		{"DeleteUnicastIpAddressEntry", &w.deleteUnicastIpAddressEntry},
		{"GetUnicastIpAddressTable", &w.getUnicastIpAddressTable},
		{"GetIpForwardTable", &w.getIpForwardTable},
		{"CreateIpForwardEntry", &w.createIpForwardEntry},
		{"DeleteIpForwardEntry", &w.deleteIpForwardEntry},
		//{"convertInterfaceIndexToLuid", &w.convertInterfaceIndexToLuid},
	}
	for _, p := range procs {
		*p.ptr, err = dll.FindProc(p.name)
		if err != nil {
			return nil, fmt.Errorf("failed to find proc %s: %w", p.name, err)
		}
	}
	return w, nil
}
