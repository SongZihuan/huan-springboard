package utils

import (
	"net"
)

func IsValidIPv4(ipString string) bool {
	ip := net.ParseIP(ipString)
	if ip == nil || ip.To4() == nil {
		return false
	}
	return true
}

func IsValidIPv6(ipString string) bool {
	ip := net.ParseIP(ipString)
	return ip != nil && ip.To4() == nil
}
