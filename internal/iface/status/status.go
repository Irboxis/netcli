package status

import (
	"github.com/spf13/cobra"
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
