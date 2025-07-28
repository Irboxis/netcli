package list

import (
	"fmt"
	"nctl/internal/utils"
	"net"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

// 存储网络接口的所有相关信息。
type InterfaceInfo struct {
	Name               string
	Status             string // "UP" 或 "DOWN"
	MACAddress         net.HardwareAddr
	MTU                int
	Flags              net.Flags
	IPAddresses        []*net.IPNet // 包含 IP 和掩码
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

			// 根据命令行参数构建目标接口名称映射
			var targetInterfaces map[string]bool
			if len(args) > 0 {
				targetInterfaces = make(map[string]bool)
				for _, arg := range args {
					targetInterfaces[arg] = true
				}
			}

			// 获取所有网络接口
			allInterfaces, err := net.Interfaces()
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "获取网络接口时出错: %v\n", err)
				return
			}

			// 处理接口并收集信息
			infos := processInterfaces(cmd, allInterfaces, targetInterfaces)

			// 如果指定了特定接口但最终没有找到任何信息，则给出提示
			if len(args) > 0 && len(infos) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "未找到指定名称的网卡：%s\n", strings.Join(args, ", "))
				return
			}

			// 打印结果
			printResults(cmd, infos, setAll)
		},
	}

	var setAll bool
	cmd.Flags().BoolVarP(&setAll, "all", "a", false, "以详细模式列出网络接口信息")

	return cmd
}

// 从单个 net.Interface 对象中提取所有相关信息
func getInterfaceInfo(cmd *cobra.Command, iface net.Interface) (*InterfaceInfo, error) {
	ip4gw, ip6gw, err := utils.GetGW(&iface)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v", err)
	}

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
			// 直接将 net.IPNet 添加到列表中
			info.IPAddresses = append(info.IPAddresses, ipNet)

			// 计算 IPv4 广播地址
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil && len(ipNet.Mask) == net.IPv4len { // 额外检查掩码长度是否为 IPv4 标准的 4 字节
				ones, bits := ipNet.Mask.Size()
				if bits == 32 && ones < 32 { // 确保是 IPv4 且不是 /32 主机路由
					broadcastIP := make(net.IP, net.IPv4len) // 确保广播 IP 的长度为 4 字节
					for i := 0; i < net.IPv4len; i++ {       // 循环迭代也应该基于 IPv4 的 4 字节长度
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

// 遍历所有网络接口，根据指定名称过滤并收集信息
func processInterfaces(cmd *cobra.Command, allInterfaces []net.Interface, targetNames map[string]bool) []InterfaceInfo {
	var infos []InterfaceInfo
	for _, iface := range allInterfaces {
		// 如果指定了网卡名称，并且当前网卡不在指定列表中，则跳过
		if targetNames != nil && !targetNames[iface.Name] {
			continue
		}

		info, err := getInterfaceInfo(cmd, iface)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "处理接口 %s 信息时出错: %v\n", iface.Name, err)
			continue
		}
		infos = append(infos, *info)
	}
	return infos
}

// 根据 detailed 标志调用不同的打印函数
func printResults(cmd *cobra.Command, infos []InterfaceInfo, detailed bool) {
	if detailed {
		printDetailed(cmd, infos)
	} else {
		printBrief(cmd, infos)
	}
}

// 以简洁模式打印网络接口信息
func printBrief(cmd *cobra.Command, infos []InterfaceInfo) {
	// 存储每列的最大可见宽度
	maxNameLen := runewidth.StringWidth("INTERFACE")
	maxStatusLen := runewidth.StringWidth("STATUS")
	maxMACLen := runewidth.StringWidth("MAC")
	maxIPLen := runewidth.StringWidth("IP")
	maxGatewayLen := runewidth.StringWidth("GATEWAY")

	// 第一遍：计算每列的最大可见宽度
	for _, info := range infos {
		if runewidth.StringWidth(info.Name) > maxNameLen {
			maxNameLen = runewidth.StringWidth(info.Name)
		}
		if runewidth.StringWidth(info.Status) > maxStatusLen {
			maxStatusLen = runewidth.StringWidth(info.Status)
		}
		if runewidth.StringWidth(info.MACAddress.String()) > maxMACLen {
			maxMACLen = runewidth.StringWidth(info.MACAddress.String())
		}

		// 找出所有IP地址中最长的CIDR字符串
		if len(info.IPAddresses) == 0 {
			if runewidth.StringWidth("N/A") > maxIPLen {
				maxIPLen = runewidth.StringWidth("N/A")
			}
		} else {
			for _, ipNet := range info.IPAddresses {
				if runewidth.StringWidth(ipNet.String()) > maxIPLen {
					maxIPLen = runewidth.StringWidth(ipNet.String())
				}
			}
		}

		// 计算网关列的最大宽度，可能包含IPv4和IPv6
		tempGatewayStrings := []string{}
		if info.DefaultGatewayIPv4 != "N/A" {
			tempGatewayStrings = append(tempGatewayStrings, info.DefaultGatewayIPv4)
		}
		if info.DefaultGatewayIPv6 != "N/A" {
			tempGatewayStrings = append(tempGatewayStrings, info.DefaultGatewayIPv6)
		}
		if len(tempGatewayStrings) == 0 { // 如果都没有，则确保 "N/A" 的宽度被考虑
			tempGatewayStrings = append(tempGatewayStrings, "N/A")
		}
		for _, gw := range tempGatewayStrings {
			if runewidth.StringWidth(gw) > maxGatewayLen {
				maxGatewayLen = runewidth.StringWidth(gw)
			}
		}
	}

	padding := 2
	nameColWidth := maxNameLen + padding
	statusColWidth := maxStatusLen + padding
	macColWidth := maxMACLen + padding
	ipColWidth := maxIPLen + padding
	gatewayColWidth := maxGatewayLen + padding

	// 打印表头
	fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s%s%s\n",
		runewidth.FillRight("INTERFACE", nameColWidth),
		runewidth.FillRight("STATUS", statusColWidth),
		runewidth.FillRight("MAC", macColWidth),
		runewidth.FillRight("IP", ipColWidth),
		runewidth.FillRight("GATEWAY", gatewayColWidth),
	)

	// 打印分隔线
	fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s%s%s\n",
		strings.Repeat("-", nameColWidth),
		strings.Repeat("-", statusColWidth),
		strings.Repeat("-", macColWidth),
		strings.Repeat("-", ipColWidth),
		strings.Repeat("-", gatewayColWidth),
	)

	// 打印数据行
	for _, info := range infos {
		var ipAddrsToPrint []string
		if len(info.IPAddresses) == 0 {
			ipAddrsToPrint = append(ipAddrsToPrint, "N/A")
		} else {
			for _, ipNet := range info.IPAddresses {
				ipAddrsToPrint = append(ipAddrsToPrint, ipNet.String())
			}
		}

		var gatewaysToPrint []string
		if info.DefaultGatewayIPv4 != "N/A" {
			gatewaysToPrint = append(gatewaysToPrint, info.DefaultGatewayIPv4)
		}
		if info.DefaultGatewayIPv6 != "N/A" {
			gatewaysToPrint = append(gatewaysToPrint, info.DefaultGatewayIPv6)
		}
		if len(gatewaysToPrint) == 0 {
			gatewaysToPrint = append(gatewaysToPrint, " ")
		}

		// brief 模式下，通常只显示第一个 IP 地址和第一个网关（如果存在）
		// 如果一个接口有多个IP，这里会打印多行，但MAC/Status/Name只打印一次
		maxRows := len(ipAddrsToPrint) // 以IP地址数量为主
		if len(gatewaysToPrint) > maxRows {
			maxRows = len(gatewaysToPrint)
		}

		for i := 0; i < maxRows; i++ {
			name := ""
			status := ""
			mac := ""

			// 只有第一行打印接口名称、状态等主信息
			if i == 0 {
				name = info.Name
				status = info.Status
				mac = info.MACAddress.String()
			}

			ipAddr := ""
			if i < len(ipAddrsToPrint) {
				ipAddr = ipAddrsToPrint[i]
			}

			gateway := ""
			if i < len(gatewaysToPrint) {
				gateway = gatewaysToPrint[i]
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s%s%s\n",
				runewidth.FillRight(name, nameColWidth),
				runewidth.FillRight(status, statusColWidth),
				runewidth.FillRight(mac, macColWidth),
				runewidth.FillRight(ipAddr, ipColWidth),
				runewidth.FillRight(gateway, gatewayColWidth),
			)
		}
	}
}

// 以表格形式显示网络接口的详细信息，并自动调整列宽。
func printDetailed(cmd *cobra.Command, infos []InterfaceInfo) {
	columnHeaders := []string{"INTERFACE", "STATUS", "MAC", "MTU", "FLAGS", "IP", "BROADCAST", "GATEWAY"}

	// 存储每列的最大可见宽度
	maxColLengths := make(map[string]int)
	for _, header := range columnHeaders {
		maxColLengths[header] = runewidth.StringWidth(header)
	}

	// 第一遍：计算每列的最大可见宽度
	for _, info := range infos {
		if runewidth.StringWidth(info.Name) > maxColLengths["INTERFACE"] {
			maxColLengths["INTERFACE"] = runewidth.StringWidth(info.Name)
		}
		if runewidth.StringWidth(info.Status) > maxColLengths["STATUS"] {
			maxColLengths["STATUS"] = runewidth.StringWidth(info.Status)
		}
		if runewidth.StringWidth(info.MACAddress.String()) > maxColLengths["MAC"] {
			maxColLengths["MAC"] = runewidth.StringWidth(info.MACAddress.String())
		}
		if runewidth.StringWidth(fmt.Sprintf("%d", info.MTU)) > maxColLengths["MTU"] {
			maxColLengths["MTU"] = runewidth.StringWidth(fmt.Sprintf("%d", info.MTU))
		}
		if runewidth.StringWidth(info.Flags.String()) > maxColLengths["FLAGS"] {
			maxColLengths["FLAGS"] = runewidth.StringWidth(info.Flags.String())
		}

		// IP 和 mask (CIDR 格式)
		if len(info.IPAddresses) == 0 {
			if runewidth.StringWidth("N/A") > maxColLengths["IP"] {
				maxColLengths["IP"] = runewidth.StringWidth("N/A")
			}
		} else {
			for _, ipNet := range info.IPAddresses {
				if runewidth.StringWidth(ipNet.String()) > maxColLengths["IP"] {
					maxColLengths["IP"] = runewidth.StringWidth(ipNet.String())
				}
			}
		}

		// 广播地址
		if len(info.BroadcastIPv4) == 0 {
			if runewidth.StringWidth("N/A") > maxColLengths["BROADCAST"] {
				maxColLengths["BROADCAST"] = runewidth.StringWidth("N/A")
			}
		} else {
			for _, bcast := range info.BroadcastIPv4 {
				if runewidth.StringWidth(bcast.String()) > maxColLengths["BROADCAST"] {
					maxColLengths["BROADCAST"] = runewidth.StringWidth(bcast.String())
				}
			}
		}

		// 默认网关
		tempGatewayStrings := []string{} // 用于宽度计算的临时切片
		if info.DefaultGatewayIPv4 != "N/A" {
			tempGatewayStrings = append(tempGatewayStrings, info.DefaultGatewayIPv4)
		}
		if info.DefaultGatewayIPv6 != "N/A" {
			tempGatewayStrings = append(tempGatewayStrings, info.DefaultGatewayIPv6)
		}
		if len(tempGatewayStrings) == 0 { // 如果都没有，则确保 "N/A" 的宽度被考虑
			tempGatewayStrings = append(tempGatewayStrings, " ")
		}
		for _, gw := range tempGatewayStrings {
			if runewidth.StringWidth(gw) > maxColLengths["GATEWAY"] {
				maxColLengths["GATEWAY"] = runewidth.StringWidth(gw)
			}
		}
	}

	padding := 2 // 列之间填充的空格数

	// 根据计算出的最大宽度和填充值，确定最终列宽
	colWidths := make(map[string]int)
	for header, length := range maxColLengths {
		colWidths[header] = length + padding
	}

	// 构建表头行和分隔线
	headerLine := ""
	separatorLine := ""
	for _, header := range columnHeaders {
		headerLine += runewidth.FillRight(header, colWidths[header])
		separatorLine += strings.Repeat("-", colWidths[header])
	}
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", headerLine)
	fmt.Fprintf(cmd.OutOrStdout(), "%s\n", separatorLine)

	// 打印数据行
	for _, info := range infos {
		var ipAddrStrings []string
		if len(info.IPAddresses) == 0 {
			ipAddrStrings = append(ipAddrStrings, "N/A")
		} else {
			for _, ipNet := range info.IPAddresses {
				ipAddrStrings = append(ipAddrStrings, ipNet.String())
			}
		}

		var broadcastStrings []string
		if len(info.BroadcastIPv4) == 0 {
			broadcastStrings = append(broadcastStrings, "N/A")
		} else {
			for _, bcast := range info.BroadcastIPv4 {
				broadcastStrings = append(broadcastStrings, bcast.String())
			}
		}

		gatewayStrings := []string{} // 重新定义，确保只包含实际网关
		if info.DefaultGatewayIPv4 != "N/A" {
			gatewayStrings = append(gatewayStrings, info.DefaultGatewayIPv4)
		}
		if info.DefaultGatewayIPv6 != "N/A" {
			gatewayStrings = append(gatewayStrings, info.DefaultGatewayIPv6)
		}
		if len(gatewayStrings) == 0 { // 如果都没有，则显示空
			gatewayStrings = append(gatewayStrings, " ")
		}

		// 找出需要打印的最大行数
		maxRows := len(ipAddrStrings)
		if len(broadcastStrings) > maxRows {
			maxRows = len(broadcastStrings)
		}
		if len(gatewayStrings) > maxRows {
			maxRows = len(gatewayStrings)
		}

		for i := 0; i < maxRows; i++ {
			name := ""
			status := ""
			mac := ""
			mtu := ""
			flags := ""

			// 只有第一行打印接口名称、状态等主信息
			if i == 0 {
				name = info.Name
				status = info.Status
				mac = info.MACAddress.String()
				mtu = fmt.Sprintf("%d", info.MTU)
				flags = info.Flags.String()
			}

			ipAddr := ""
			if i < len(ipAddrStrings) {
				ipAddr = ipAddrStrings[i]
			}

			broadcast := ""
			if i < len(broadcastStrings) {
				broadcast = broadcastStrings[i]
			}

			gateway := ""
			if i < len(gatewayStrings) {
				gateway = gatewayStrings[i]
			}

			// 拼接并打印当前行
			fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s%s%s%s%s%s\n",
				runewidth.FillRight(name, colWidths["INTERFACE"]),
				runewidth.FillRight(status, colWidths["STATUS"]),
				runewidth.FillRight(mac, colWidths["MAC"]),
				runewidth.FillRight(mtu, colWidths["MTU"]),
				runewidth.FillRight(flags, colWidths["FLAGS"]),
				runewidth.FillRight(ipAddr, colWidths["IP"]),
				runewidth.FillRight(broadcast, colWidths["BROADCAST"]),
				runewidth.FillRight(gateway, colWidths["GATEWAY"]),
			)
		}
	}
}
