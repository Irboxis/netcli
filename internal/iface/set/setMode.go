package set

import "github.com/spf13/cobra"

var (
	setMode    string
	setPromisc bool
	setDuplex  string
	setOnly    string
)

func setModes(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVarP(&setMode, "mode", "M", "", "Network interface working mode (value: dhcp, local, ip)")
	cmd.Flags().BoolVarP(&setPromisc, "promisc", "p", false, "Whether to enable promiscuous mode")
	cmd.Flags().StringVarP(&setDuplex, "duplex", "x", "", "Full/half duplex mode of the network (value: half, full)")
	cmd.Flags().StringVarP(&setOnly, "only", "o", "", "Use only IPv4 or IPv6 (value: 4, 6)")

	return cmd
}
