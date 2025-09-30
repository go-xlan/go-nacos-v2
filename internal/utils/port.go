package utils

import (
	"regexp"
	"strings"

	"github.com/yyle88/must"
)

// addressRegexp matches IPv4:port format addresses
// Captures IP octets and port number in separate groups
//
// addressRegexp 匹配 IPv4:端口格式的地址
// 在不同组中捕获 IP 八位组和端口号
var addressRegexp = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+):(\d+)$`)

// MustGetIPv4 extracts IPv4 address from address string
// Parses IP:port format and returns IP part
// Panics if address format is invalid
//
// MustGetIPv4 从地址字符串提取 IPv4 地址
// 解析 IP:端口格式并返回 IP 部分
// 地址格式无效时会 panic
func MustGetIPv4(address string) string {
	parts := addressRegexp.FindStringSubmatch(address)
	must.Len(parts, 6)
	return strings.Join(parts[1:5], ".") // First 4 capture groups are IP octets // 前 4 个捕获组是 IP 八位组
}

// MustGetPort extracts port number from address string
// Parses IP:port format and returns port part
// Panics if address format is invalid
//
// MustGetPort 从地址字符串提取端口号
// 解析 IP:端口格式并返回端口部分
// 地址格式无效时会 panic
func MustGetPort(address string) string {
	parts := addressRegexp.FindStringSubmatch(address)
	must.Len(parts, 6)
	return parts[5] // 5th capture group is port number // 第 5 个捕获组是端口号
}
