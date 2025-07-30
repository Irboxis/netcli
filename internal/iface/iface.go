package iface

import (
	"nctl/internal/iface/list"
	"nctl/internal/iface/status"

	"github.com/spf13/cobra"
)

var ifaceCmd = &cobra.Command{
	Use:   "iface",
	Short: "Host network interface management",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// 注册所有iface 下的子命令
func RegisterIfaceCommands(rootCmd *cobra.Command) {
	// 挂载 iface 子命令
	rootCmd.AddCommand(ifaceCmd)

	// 挂载 iface list 命令
	ifaceCmd.AddCommand(list.List())
	// 挂载 iface status 系列命令
	ifaceCmd.AddCommand(status.Status())
}
