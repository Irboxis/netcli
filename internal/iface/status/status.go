package status

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
)

func Status() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Network interface state management",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(up())
	cmd.AddCommand(down())

	return cmd
}

func up() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Open Network Connections",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Fprintf(cmd.OutOrStderr(), "至少存在一个网络接口\n")
				cmd.Help()
				return
			}

			run(args, true)
		},
	}

	return cmd
}

func down() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Shut down the network interface",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Fprintf(cmd.OutOrStderr(), "至少存在一个网络接口\n")
				cmd.Help()
				return
			}

			run(args, false)
		},
	}

	return cmd
}

func run(ifacesName []string, enable bool) {
	action := "DOWN"
	if enable {
		action = "UP"
	}

	for _, name := range ifacesName {
		var err error
		switch runtime.GOOS {
		case "linux":
			err = toggleLinuxInterface(name, enable)
		case "windows":
			err = toggleWindowsInterface(name, enable)
		default:
			fmt.Printf("操作系统 %s 不支持对网络接口 %s 的 %s 操作。\n", runtime.GOOS, name, action)
			continue
		}

		if err != nil {
			fmt.Printf("对网络接口 %s 进行 %s 操作失败: %v\n", name, action, err)
		} else {
			fmt.Printf("网络接口 %s 已成功 %s。\n", name, action)
		}
	}
}

// Linux 上操作网络接口
func toggleLinuxInterface(name string, enable bool) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("查找网络接口 %s 失败: %v", name, err)
	}

	if enable {
		err = netlink.LinkSetUp(link)
	} else {
		err = netlink.LinkSetDown(link)
	}

	if err != nil {
		return fmt.Errorf("设置网络接口 %s 状态失败: %v", name, err)
	}
	return nil
}

// Windows 上操作网络接口
func toggleWindowsInterface(name string, enable bool) error {
	actionCmd := "DISABLED"
	if enable {
		actionCmd = "ENABLED"
	}

	// 注意：在 Windows 上执行此命令通常需要管理员权限。
	cmd := exec.Command("netsh", "interface", "set", "interface", name, actionCmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行 netsh 命令失败: %v, 输出: %s", err, string(output))
	}
	return nil
}
