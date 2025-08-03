//go: build linux

package linux

import "nctl/internal/utils"

type UnixNctl struct{}

// 工厂函数
func NewNctltools() utils.Ifaces {
	return &UnixNctl{}
}
