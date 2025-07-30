package list

import (
	"fmt"
	"net"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"nctl/internal/utils"
)

// InterfaceInfo 存储网络接口的所有相关信息。
type InterfaceInfo struct {
	Name               string
	Status             string
	MACAddress         net.HardwareAddr
	MTU                int
	Flags              net.Flags
	IPAddresses        []*net.IPNet
	BroadcastIPv4      []net.IP
	DefaultGatewayIPv4 string
	DefaultGatewayIPv6 string
}

func List() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [interface_name...]",
		Short: "列出指定或所有网络接口信息",
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			setAll, _ := cmd.Flags().GetBool("all")

			var targetInterfaces map[string]bool
			if len(args) > 0 {
				targetInterfaces = make(map[string]bool)
				for _, arg := range args {
					targetInterfaces[arg] = true
				}
			}

			allInterfaces, err := net.Interfaces()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "获取网络接口时出错: %v\n", err)
				return
			}

			infos := processInterfaces(cmd, allInterfaces, targetInterfaces)

			if len(args) > 0 && len(infos) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "未找到指定名称的网卡：%s\n", strings.Join(args, ", "))
				return
			}

			printResults(cmd, infos, setAll)
		},
	}

	cmd.Flags().BoolP("all", "a", false, "以详细模式列出网络接口信息")

	return cmd
}

func getInterfaceInfo(iface net.Interface) (*InterfaceInfo, error) {
	ip4gw, ip6gw := utils.GetGW(&iface)

	info := &InterfaceInfo{
		Name:               iface.Name,
		MACAddress:         iface.HardwareAddr,
		MTU:                iface.MTU,
		Flags:              iface.Flags,
		DefaultGatewayIPv4: ip4gw,
		DefaultGatewayIPv6: ip6gw,
	}

	if iface.Flags&net.FlagUp != 0 {
		info.Status = "UP"
	} else {
		info.Status = "DOWN"
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("获取接口 %s 的地址时出错: %v", iface.Name, err)
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			info.IPAddresses = append(info.IPAddresses, ipNet)

			ipv4 := ipNet.IP.To4()
			if ipv4 != nil && len(ipNet.Mask) == net.IPv4len {
				ones, bits := ipNet.Mask.Size()
				if bits == 32 && ones < 32 {
					broadcastIP := make(net.IP, net.IPv4len)
					for i := 0; i < net.IPv4len; i++ {
						broadcastIP[i] = ipv4[i] | (^ipNet.Mask[i])
					}
					if !broadcastIP.Equal(ipv4) && !broadcastIP.IsUnspecified() {
						info.BroadcastIPv4 = append(info.BroadcastIPv4, broadcastIP)
					}
				}
			}
		}
	}
	return info, nil
}

func processInterfaces(cmd *cobra.Command, allInterfaces []net.Interface, targetNames map[string]bool) []InterfaceInfo {
	var infos []InterfaceInfo
	for _, iface := range allInterfaces {
		if targetNames != nil && !targetNames[iface.Name] {
			continue
		}
		info, err := getInterfaceInfo(iface)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "处理接口 %s 信息时出错: %v\n", iface.Name, err)
			continue
		}
		infos = append(infos, *info)
	}
	return infos
}

func printResults(cmd *cobra.Command, infos []InterfaceInfo, detailed bool) {
	if detailed {
		printDetailed(cmd, infos)
	} else {
		printBrief(cmd, infos)
	}
}

func printBrief(cmd *cobra.Command, infos []InterfaceInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.AppendHeader(table.Row{"INTERFACE", "STATUS", "MAC", "IP", "GATEWAY"})

	for _, info := range infos {
		ipAddrs := toStringSlice(info.IPAddresses)
		if len(ipAddrs) == 0 {
			ipAddrs = []string{"N/A"}
		}

		gateways := gatherGateways(info)
		if len(gateways) == 0 {
			gateways = []string{" "}
		}

		max := maxLen(ipAddrs, gateways)
		for i := 0; i < max; i++ {
			name, status, mac := "", "", ""
			if i == 0 {
				name = info.Name
				status = info.Status
				mac = info.MACAddress.String()
			}
			t.AppendRow([]interface{}{
				name,
				status,
				mac,
				getSafe(ipAddrs, i),
				getSafe(gateways, i),
			})
		}
	}
	t.Render()
}

func printDetailed(cmd *cobra.Command, infos []InterfaceInfo) {
	t := table.NewWriter()
	t.SetOutputMirror(cmd.OutOrStdout())
	t.AppendHeader(table.Row{
		"INTERFACE", "STATUS", "MAC", "MTU", "FLAGS", "IP", "BROADCAST", "GATEWAY",
	})

	for _, info := range infos {
		ipAddrs := toStringSlice(info.IPAddresses)
		if len(ipAddrs) == 0 {
			ipAddrs = []string{"N/A"}
		}
		bcasts := toStringSlice(info.BroadcastIPv4)
		if len(bcasts) == 0 {
			bcasts = []string{"N/A"}
		}
		gateways := gatherGateways(info)
		if len(gateways) == 0 {
			gateways = []string{" "}
		}

		max := maxLen(ipAddrs, bcasts, gateways)
		for i := 0; i < max; i++ {
			name, status, mac, mtu, flags := "", "", "", "", ""
			if i == 0 {
				name = info.Name
				status = info.Status
				mac = info.MACAddress.String()
				mtu = fmt.Sprintf("%d", info.MTU)
				flags = info.Flags.String()
			}
			t.AppendRow([]interface{}{
				name,
				status,
				mac,
				mtu,
				flags,
				getSafe(ipAddrs, i),
				getSafe(bcasts, i),
				getSafe(gateways, i),
			})
		}
	}
	t.Render()
}

func toStringSlice[T fmt.Stringer](list []T) []string {
	var result []string
	for _, item := range list {
		result = append(result, item.String())
	}
	return result
}

func gatherGateways(info InterfaceInfo) []string {
	var gw []string
	if info.DefaultGatewayIPv4 != "N/A" {
		gw = append(gw, info.DefaultGatewayIPv4)
	}
	if info.DefaultGatewayIPv6 != "N/A" {
		gw = append(gw, info.DefaultGatewayIPv6)
	}
	return gw
}

func getSafe(list []string, idx int) string {
	if idx < len(list) {
		return list[idx]
	}
	return ""
}

func maxLen(lists ...[]string) int {
	max := 0
	for _, list := range lists {
		if len(list) > max {
			max = len(list)
		}
	}
	return max
}
