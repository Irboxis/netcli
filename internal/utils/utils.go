package utils

type Ifaces interface {
	// 检查接口存在性
	IsExistingIface(iface string) error
}

type WindowsNctl struct{}
type UnixNctl struct{}

// 统一返回各个文件的工厂函数
// 所有需要实现 Ifaces 接口和 xxxNctl 结构体的文件，都必须添加 newNctltools 函数
func NewNctlUtils() Ifaces {
	return newNctltools()
}
