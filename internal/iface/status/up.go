package status

import (
	"fmt"
	"os"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

func up() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Open Network Connections",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Fprintf(cmd.OutOrStderr(), "至少存在一个网络接口")
				cmd.Help()
				return
			}

			runUP(args)
		},
	}

	return cmd
}

func runUP(ifaceName []string) {
	switch runtime.GOOS {
	case "linux":
		linuxUp(ifaceName)
	case "windows":
		windowsUp(ifaceName)
	case "darwin":
		macosUp(ifaceName)
	default:
		fmt.Fprintf(os.Stderr, "未知的操作系统")
	}
}

func linuxUp(ifaceName []string) {
	for _, iface := range ifaceName {
		link, err := netlink.LinkByName(iface)
		if err != nil {
			fmt.Fprintf(os.Stderr, "意料之外的网络接口: %v", err)
			continue
		}

		if err := netlink.LinkSetUp(link); err != nil {
			fmt.Fprintf(os.Stderr, "打开接口 '%s' 失败：%v", iface, err)
		} else {
			fmt.Fprintf(os.Stdout, "打开接口 '%s' 成功", iface)
		}
	}
}

// Windows端操作函数，使用WMI实现
func windowsUp(ifaceName []string) {
	// 初始化 COM
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil && err != syscall.S_FALSE {

	}
}

func macosUp(ifaceName []string) {

}
