//go:build windows

package utils

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

// 工厂函数
func newNctltools() Ifaces {
	return &WindowsNctl{}
}

// 检查接口名称的存在性
func (w *WindowsNctl) IsExistingIface(iface string) error {
	bufLen := uint32(15 * 1024) // 初始 buffer 大小：15KB
	for attempt := 0; attempt < 2; attempt++ {
		buf := make([]byte, bufLen)
		pAdapter := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))

		ret := windows.GetAdaptersAddresses(
			windows.AF_UNSPEC,
			windows.GAA_FLAG_INCLUDE_ALL_INTERFACES,
			0,
			pAdapter,
			&bufLen,
		)

		if ret == windows.ERROR_BUFFER_OVERFLOW {
			// buffer 不够大，下一轮尝试更大的 buffer
			continue
		} else if ret != windows.ERROR_SUCCESS {
			return fmt.Errorf("GetAdaptersAddresses failed, code=%v", ret)
		}

		// 成功获取信息，开始遍历适配器
		for aa := pAdapter; aa != nil; aa = aa.Next {
			friendlyName := windows.UTF16PtrToString(aa.FriendlyName)
			if friendlyName == iface {
				return nil
			}
		}

		return fmt.Errorf("interface %s not found", iface)
	}

	return fmt.Errorf("unable to allocate sufficient buffer for GetAdaptersAddresses")
}

// 编译时接口检查
var _ Ifaces = (*WindowsNctl)(nil)
