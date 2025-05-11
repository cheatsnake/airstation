package network

import (
	"net"
	"strings"
)

func IsLocalhost(host string) bool {
	hostWithoutPort := strings.Split(host, ":")[0]

	if hostWithoutPort == "localhost" || hostWithoutPort == "127.0.0.1" || hostWithoutPort == "::1" {
		return true
	}

	ip := net.ParseIP(hostWithoutPort)
	if ip != nil {
		privateIPBlocks := []net.IPNet{
			{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
			{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
			{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		}

		for _, block := range privateIPBlocks {
			if block.Contains(ip) {
				return true
			}
		}
	}

	return false
}
