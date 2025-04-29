package utils

import (
	"net"

	"github.com/yyle88/erero"
)

func GetIpv4() (string, error) {
	nets := map[string]bool{
		"en0":    true,
		"eth0":   true,
		"ens224": true,
		"ens5":   true,
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", erero.Wro(err)
	}

	for _, item := range interfaces {
		if item.Flags&net.FlagUp != net.FlagUp {
			continue
		}
		if item.Flags&net.FlagLoopback == net.FlagLoopback {
			continue
		}
		addresses, err := item.Addrs()
		if err != nil {
			continue // Skip this unknown interface
		}

		// Check if the current interface is the specified IP-address
		if nets[item.Name] {
			for _, address := range addresses {
				if ipNet, ok := address.(*net.IPNet); ok && ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
					return ipNet.IP.String(), nil
				}
			}
		}
	}

	return "", erero.New("没有从本地网卡找到ipv4")
}
