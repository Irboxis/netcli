//go:build windows

package windows

import (
	"fmt"
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
