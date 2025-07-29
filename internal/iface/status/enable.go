package status

import "github.com/spf13/cobra"

func enable() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Enable the network interface",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
}
