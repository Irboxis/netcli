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

			RunStatus(args, true)
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

			RunStatus(args, false)
		},
	}

	return cmd
}

func RunStatus(ifacesName []string, enable bool) {
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
			fmt.Printf("OS %s does not support the %s operation on the interface %s\n", runtime.GOOS, name, action)
			continue
		}

		if err != nil {
			fmt.Printf("operation %s failed on interface %s: %v\n", name, action, err)
		} else {
			fmt.Printf("interface %s has successfully %s\n", name, action)
		}
	}
}

// Linux 上操作网络接口
func toggleLinuxInterface(name string, enable bool) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return fmt.Errorf("failed to find network interface %s: %v", name, err)
	}

	if enable {
		err = netlink.LinkSetUp(link)
	} else {
		err = netlink.LinkSetDown(link)
	}

	if err != nil {
		return fmt.Errorf("failed to set network interface %s status: %v", name, err)
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
