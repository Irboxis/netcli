// go: build windows

package windows

import (
	"nctl/internal/utils"

	"golang.org/x/sys/windows"
)

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
	PrefixOrigin       uint32 // Nlpo...
	SuffixOrigin       uint32 // Nlso...
	ValidLifetime      uint32
	PreferredLifetime  uint32
	OnLinkPrefixLength uint8
	SkipAsSource       bool
	DadState           uint32 // Nlds...
	ScopeId            uint32
	CreationTimeStamp  int64
}

// 工厂函数
func NewNctltools() utils.Ifaces {
	return &WindowsNctl{}
}

// 编译时接口检查
var _ utils.Ifaces = (*WindowsNctl)(nil)
