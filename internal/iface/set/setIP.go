package set

import (
	"fmt"
	"nctl/internal/utils"
	"os"

	"github.com/spf13/cobra"
)

var (
	setIP  []string
	setDNS []string
	setADD bool
	setDEL bool
	setGW  string
)

func setAddrs(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().StringSliceVarP(&setIP, "ip", "i", []string{}, "Accepts one or more addresses (ipv4 or ipv6)")
	cmd.Flags().StringSliceVarP(&setDNS, "dns", "d", []string{}, "Accepts one or more DNS addresses (ipv4 or ipv6)")
	cmd.Flags().BoolVarP(&setADD, "add", "I", false, "Add address, used with --ip or --dns")
	cmd.Flags().BoolVarP(&setDEL, "del", "S", false, "Delete address, used with --ip or --dns")
	cmd.Flags().StringVarP(&setGW, "gw", "g", "", "Set the default gateway")

	return cmd
}

func runAddrs(ifaceName string) {
	ifaceUtils := utils.NewNctlUtils()
	if err := ifaceUtils.IsExistingIface(ifaceName); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

}
