package status

import (
	"github.com/spf13/cobra"
)

func RegisterStatusCommands(rootCmd *cobra.Command) {
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Network interface state management",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	statusCmd.AddCommand(up())
	statusCmd.AddCommand(down())
	statusCmd.AddCommand(disable())
	statusCmd.AddCommand(enable())
}
