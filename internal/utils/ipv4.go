// Package utils: Internal utilities for network address detection and parsing
// Provides IPv4 address extraction from network interfaces and port parsing
//
// utils: 网络地址检测和解析的内部工具
// 提供从网络接口提取 IPv4 地址和端口解析功能
package utils

import (
	"net"

	"github.com/pkg/errors"
	"github.com/yyle88/erero"
)

// allowedInterfaces defines permitted network interface names
// Restricts IP detection to standard network interfaces
//
// allowedInterfaces 定义允许的网络接口名称
// 限制 IP 检测到标准网络接口
var allowedInterfaces = map[string]struct{}{
	"en0":    {}, // macOS Ethernet/Wi-Fi // macOS 以太网/Wi-Fi
	"eth0":   {}, // Linux Ethernet // Linux 以太网
	"ens224": {}, // VMware virtual network // VMware 虚拟网络
	"ens5":   {}, // AWS EC2 network // AWS EC2 网络
}

// GetIPv4 retrieves IPv4 address from allowed network interfaces
// Scans system interfaces and returns first valid IPv4 address
// Returns error if no valid address found
//
// GetIPv4 从允许的网络接口获取 IPv4 地址
// 扫描系统接口并返回第一个有效的 IPv4 地址
// 未找到有效地址时返回错误
func GetIPv4() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", erero.Wro(err)
	}
	return GetIPv4FromInterfaces(interfaces, allowedInterfaces)
}

// GetIPv4FromInterfaces extracts IPv4 from specified network interfaces
// Filters interfaces by name and checks for valid non-loopback IPv4 addresses
// Returns first matching address or error with details
//
// GetIPv4FromInterfaces 从指定的网络接口提取 IPv4
// 按名称过滤接口并检查有效的非回环 IPv4 地址
// 返回第一个匹配地址或带详细信息的错误
func GetIPv4FromInterfaces(interfaces []net.Interface, allowedInterfaceNames map[string]struct{}) (string, error) {
	var errs []error
	for _, ifc := range interfaces {
		if ifc.Flags&net.FlagUp != net.FlagUp {
			continue
		}
		if ifc.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}
		if _, ok := allowedInterfaceNames[ifc.Name]; !ok {
			continue
		}

		addresses, err := ifc.Addrs()
		if err != nil {
			errs = append(errs, errors.WithMessagef(err, "unable to get addresses on %s", ifc.Name))
			continue
		}

		for _, address := range addresses {
			if ipNet, ok := address.(*net.IPNet); ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
				return ipNet.IP.String(), nil
			}
		}
	}

	if len(errs) > 0 {
		return "", erero.Joins(errs)
	}
	return "", erero.New("no IPv4 address found on allowed network interfaces")
}
