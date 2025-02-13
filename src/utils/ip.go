package utils

import "net"

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

func IsPrivateIP(ipString string, includeLookback bool) bool {
	ip := net.ParseIP(ipString)
	if ip == nil {
		return false
	}

	if includeLookback && ip.IsLoopback() {
		return true
	}

	// Check for IPv4 private addresses
	for _, privIPv4Net := range []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
	} {
		_, privIPv4Range, _ := net.ParseCIDR(privIPv4Net)
		if privIPv4Range.Contains(ip) {
			return true
		}
	}

	// Check for IPv6 unique local address (ULA)
	ulaIPv6Net := "fc00::/7"
	_, ulaIPv6Range, _ := net.ParseCIDR(ulaIPv6Net)
	if ip.To4() == nil && ulaIPv6Range.Contains(ip) {
		return true
	}

	return false
}

func IsLoopbackIP(ipString string) bool {
	ip := net.ParseIP(ipString)
	return ip != nil && (ip.IsLoopback())
}
