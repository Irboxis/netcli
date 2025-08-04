// package set

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
	cmd.Flags().StringSliceVarP(&setIP, "ip", "i", []string{}, "Set one or more IP addresses in CIDR format (e.g., 192.168.1.10/24)")
	cmd.Flags().StringSliceVarP(&setDNS, "dns", "d", []string{}, "Set one or more DNS server addresses")
	cmd.Flags().BoolVarP(&setADD, "add", "I", false, "Add IP/DNS to the interface instead of overwriting")
	cmd.Flags().BoolVarP(&setDEL, "del", "S", false, "Delete IP/DNS from the interface")
	cmd.Flags().StringVarP(&setGW, "gw", "g", "", "Set the default gateway (always overwrites)")

	return cmd
}

// 关键字格式和数据合法性检查
func checkAddrs(cmd *cobra.Command) bool {
	if setADD && setDEL {
		fmt.Fprintln(os.Stderr, "Error: The --add and --del flags cannot be used together.")
		cmd.Help()
		return false
	}

	// 如果指定了 --add 或 --del，但没有提供 --ip 或 --dns，则为错误
	if (setADD || setDEL) && len(setIP) == 0 && len(setDNS) == 0 {
		fmt.Fprintln(os.Stderr, "Error: The --add or --del flag must be used with --ip or --dns.")
		cmd.Help()
		return false
	}
	return true
}

// 将字符串解析为 *net.IPNet
func parseIPs(ipStrs []string) ([]*net.IPNet, error) {
	var ipnets []*net.IPNet
	for _, ipStr := range ipStrs {
		ip, ipnet, err := net.ParseCIDR(ipStr)
		if err != nil {
			// 尝试将其作为没有掩码的普通 IP 地址进行解析，并赋予一个默认掩码
			// IPv4 使用 /32, IPv6 使用 /128
			parsedIP := net.ParseIP(ipStr)
			if parsedIP == nil {
				return nil, fmt.Errorf("invalid IP address format: %s", ipStr)
			}
			if parsedIP.To4() != nil {
				ipStr = ipStr + "/32"
			} else {
				ipStr = ipStr + "/128"
			}
			ip, ipnet, err = net.ParseCIDR(ipStr)
			if err != nil {
				// 这不应该发生
				return nil, fmt.Errorf("failed to parse IP %s with default mask", ipStr)
			}
		}
		ipnet.IP = ip // 确保 IPNet 包含解析出的 IP
		ipnets = append(ipnets, ipnet)
	}
	return ipnets, nil
}

// 将字符串解析为 net.IP
func parseDNSs(dnsStrs []string) ([]net.IP, error) {
	var ips []net.IP
	for _, s := range dnsStrs {
		ip := net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf("invalid DNS address format: %s", s)
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

func runAddrs(ifaceName string, cmd *cobra.Command) {
	if !checkAddrs(cmd) {
		return
	}

	ifaceUtils := utils.IfaceUtils()
	if err := ifaceUtils.IsExistingIface(ifaceName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	// 处理 ip
	if len(setIP) > 0 {
		ipnets, err := parseIPs(setIP)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing IP addresses: %v\n", err)
			return
		}

		if setADD {
			fmt.Printf("Adding IP addresses to %s...\n", ifaceName)
			for _, ipnet := range ipnets {
				if err := ifaceUtils.AddIP(ifaceName, ipnet); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add IP %s: %v\n", ipnet.String(), err)
				}
			}
		} else if setDEL {
			fmt.Printf("Deleting IP addresses from %s...\n", ifaceName)
			for _, ipnet := range ipnets {
				if err := ifaceUtils.DelIP(ifaceName, ipnet); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to delete IP %s: %v\n", ipnet.String(), err)
				}
			}
		} else { // 默认：覆盖
			fmt.Printf("Overwriting IP addresses on %s...\n", ifaceName)
			if err := ifaceUtils.SetIPs(ifaceName, ipnets); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to set IPs: %v\n", err)
			}
		}
	}

	// 处理 dns
	if len(setDNS) > 0 {
		dnsIPs, err := parseDNSs(setDNS)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing DNS addresses: %v\n", err)
			return
		}

		if setADD {
			fmt.Printf("Adding DNS servers to %s...\n", ifaceName)
			for _, ip := range dnsIPs {
				if err := ifaceUtils.AddDNS(ifaceName, ip); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to add DNS %s: %v\n", ip.String(), err)
				}
			}
		} else if setDEL {
			fmt.Printf("Deleting DNS servers from %s...\n", ifaceName)
			for _, ip := range dnsIPs {
				if err := ifaceUtils.DelDNS(ifaceName, ip); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to delete DNS %s: %v\n", ip.String(), err)
				}
			}
		} else { // 默认：覆盖
			fmt.Printf("Overwriting DNS servers on %s...\n", ifaceName)
			if err := ifaceUtils.SetDNSs(ifaceName, dnsIPs); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to set DNS servers: %v\n", err)
			}
		}
	}

	// 处理网关 (总是覆盖)
	if setGW != "" {
		ip := net.ParseIP(setGW)
		if ip == nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid gateway address format: '%s'\n", setGW)
			return
		}

		fmt.Printf("Setting default gateway on %s...\n", ifaceName)
		if err := ifaceUtils.SetGateway(ifaceName, ip); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting gateway %s: %v\n", ip.String(), err)
		}
	}
}
