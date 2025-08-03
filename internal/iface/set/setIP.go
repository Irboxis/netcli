package set

import (
	"fmt"
	"nctl/internal/utils"
	"net"
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

// 关键字格式和数据合法性检查
func checkAddrs(cmd *cobra.Command) {
	if setADD && setDEL {
		fmt.Fprintf(os.Stderr, "The --add and --del flags cannot be used together")
		cmd.Help()
		return
	}

	configAddrCount := len(setIP) > 0 || len(setDNS) > 0 || setGW != ""
	if configAddrCount {
		if (setADD || setDEL) && len(setIP) == 0 && len(setDNS) == 0 {
			fmt.Fprintf(os.Stderr, "The --add or --del flag must be used with --ip or --dns\n")
			cmd.Help()
			return
		}
	}
}

// 核心逻辑
func runAddrs(ifaceName string, cmd *cobra.Command) {
	checkAddrs(cmd)

	// 检查接口合法性
	ifaceUtils := utils.NewNctlUtils()
	if err := ifaceUtils.IsExistingIface(ifaceName); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	var actionToken bool
	var errors []error
	if len(setIP) > 0 {
		actionToken = true
		var ipActionFunc func(string, *net.IPNet) error
		actionGerund := ""

		if setADD {
			actionGerund = "add"
			ipActionFunc = ifaceUtils.AddIP
		} else if setDEL {
			actionGerund = "delete"
			ipActionFunc = ifaceUtils.DelIP
		} else {
			actionGerund = "set"
		}

		for _, ipstr := range setIP {
			_, ipnet, err := net.ParseCIDR(ipstr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: Invalid IP address format for '%s'. Please use CIDR notation (e.g., 192.168.1.1/24)\n", ipstr)
				errors = append(errors, err)
				continue
			}

			if ipActionFunc != nil {
				// 添加或删除操作
				if err := ipActionFunc(ifaceName, ipnet); err != nil {
					errMessage := fmt.Errorf("failed to %s %s: %w", actionGerund, ipnet.String(), err)
					fmt.Fprintln(os.Stderr, errMessage)
					errors = append(errors, errMessage)
				}
			} else {
				// 执行覆盖操作（先删除，再添加）
				_ = ifaceUtils.DelIP(ifaceName, ipnet)
				if err := ifaceUtils.AddIP(ifaceName, ipnet); err != nil {
					errMessage := fmt.Errorf("failed to %s %s: %w", actionGerund, ipnet.String(), err)
					fmt.Fprintln(os.Stderr, errMessage)
					errors = append(errors, errMessage)
				}
			}
		}
	}

	if len(setDNS) > 0 {
		var parsedDNS []net.IP
		for _, dnsstr := range setDNS {
			ip := net.ParseIP(dnsstr)
			if ip == nil {
				fmt.Fprintf(os.Stderr, "Invalid DNS address format for '%s'\n", dnsstr)
				return
			}
			parsedDNS = append(parsedDNS, ip)
		}

		actionToken = true
		// 对接口的 dns 地址进行操作
		if setADD {

		} else if setDEL {

		} else {

		}
	}

	if setGW != "" {
		ip := net.ParseIP(setGW)
		if ip == nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid Gateway address format for '%s'\n", setGW)
			return
		}

		actionToken = true
		// 对接口的 gateway 配置进行操作

	}

	if actionToken {
		fmt.Fprintf(os.Stdout, "Update interface '%s' configuration successfully\n", ifaceName)
	}
}
