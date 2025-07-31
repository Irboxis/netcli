package set

import "github.com/spf13/cobra"

var (
	setMAC   string
	setMTU   int
	setARP   bool
	setQOS   int
	setVlan  string
	setDesc  string
	setTTL   int
	setSpeed string
)

func setOthers(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringVarP(&setMAC, "mac", "m", "", "Network interface MAC address")
	cmd.Flags().IntVarP(&setMTU, "mtu", "u", 0, "Set mtu value")
	cmd.Flags().BoolVarP(&setARP, "arp", "a", false, "Whether to enable ARP")
	cmd.Flags().IntVarP(&setQOS, "qos", "q", 0, "Set QOS value")
	cmd.Flags().StringVarP(&setVlan, "vlan", "v", "", "Set VLAN tag of interface")
	cmd.Flags().StringVarP(&setDesc, "desc", "c", "", "Set description of interface")
	cmd.Flags().IntVarP(&setTTL, "ttl", "t", 0, "Set TTL value")
	cmd.Flags().StringVarP(&setSpeed, "speed", "s", "", "Set interface network rate, supporting units of B, K, M, G")

	return cmd
}
