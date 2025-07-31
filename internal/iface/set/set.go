package set

import (
	"fmt"
	"nctl/internal/iface/status"
	"nctl/internal/utils"
	"os"

	"github.com/spf13/cobra"
)

var (
	setUp    bool
	setDown  bool
	setReset bool
)

func SetC() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Editing Network Interface Details",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ifaceName := args[0]

			flagCount := 0
			if setUp {
				flagCount++
			}
			if setDown {
				flagCount++
			}
			if setReset {
				flagCount++
			}

			// 检查是否存在状态关键字，并检查唯一性
			if flagCount > 1 {
				fmt.Fprintf(os.Stderr, "There can only be one state management (UP, DOWN, RESET)\n")
				return
			}

			// 检查是否存在 setIP 中规定的关键字
			configAddrCount := len(setIP) > 0 || len(setDNS) > 0 || setGW != ""

			if flagCount == 0 && !configAddrCount {
				cmd.Help()
				return
			}

			if configAddrCount {
				// setadd 或 serdel 标志必须搭配 setip 或 setdns 使用
				if (setADD || setDEL) && len(setIP) == 0 && len(setDNS) == 0 {
					fmt.Fprintf(os.Stderr, "The --add or --del flag must be used with --ip or --dns\n")
					cmd.Help()
					return
				}

				runAddrs(ifaceName)
			}

			if setUp {
				RunSet(ifaceName, true)
			} else if setDown {
				RunSet(ifaceName, false)
			} else if setReset {
				setResetFunc(ifaceName)
			}
		},
	}

	// 状态设置
	cmd.Flags().BoolVarP(&setUp, "up", "U", false, "Open the network interface")
	cmd.Flags().BoolVarP(&setDown, "down", "D", false, "Shut down the network interface")
	cmd.Flags().BoolVarP(&setReset, "reset", "R", false, "Reset network interface configuration")
	// ip相关设置
	cmd = setAddrs(cmd)
	// 模式相关设置
	cmd = setModes(cmd)
	// 其他设置
	cmd = setOthers(cmd)

	return cmd
}

func RunSet(ifaceName string, enable bool) {
	ifaceUtils := utils.NewNctlUtils()
	if err := ifaceUtils.IsExistingIface(ifaceName); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// 处理 up 和 down 关键字，直接调用 status 命令的逻辑即可
	status.RunStatus([]string{ifaceName}, enable)
}

// 处理 reset 关键字
func setResetFunc(ifaceName string) {

}
