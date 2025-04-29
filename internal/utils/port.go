package utils

import (
	"regexp"
	"strings"

	"github.com/yyle88/must"
)

func MustGetIpV4(address string) string {
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+):(\d+)$`)
	matches := re.FindStringSubmatch(address)
	must.Len(matches, 6)
	return strings.Join(matches[1:5], ".") // 第 5 个捕获组是端口
}

func MustGetPort(address string) string {
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+):(\d+)$`)
	matches := re.FindStringSubmatch(address)
	must.Len(matches, 6)
	return matches[5] // 第 5 个捕获组是端口
}
