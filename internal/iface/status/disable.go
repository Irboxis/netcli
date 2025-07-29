package status

import "github.com/spf13/cobra"

func disable() *cobra.Command {
	return &cobra.Command{
		Use:   "disable",
		Short: "Disable a network interface",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
}
