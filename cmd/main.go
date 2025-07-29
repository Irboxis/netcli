package main

import (
	"fmt"
	"nctl/internal/iface"

	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "nctl",
		Short: "A comprehensive network CLI tool",
		Long:  "net is a powerful command-line interface for network diagnostics, configuration, and management",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// 挂载 iface 系列命令
	iface.RegisterIfaceCommands(rootCmd)

	// 执行根命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("hello")
}
