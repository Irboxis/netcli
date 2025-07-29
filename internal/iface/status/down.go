package status

import "github.com/spf13/cobra"

func down() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Shut down the network interface",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	return cmd
}
