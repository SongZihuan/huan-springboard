package utils

import (
	"net"
	"strings"
)

func IsValidIPv4CIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil && !strings.Contains(cidr, ":")
}

// IsValidIPv6CIDR checks if the given string is a valid IPv6 CIDR notation.
func IsValidIPv6CIDR(cidr string) bool {
	_, _, err := net.ParseCIDR(cidr)
	return err == nil && strings.Contains(cidr, ":")
}
