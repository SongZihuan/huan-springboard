package ipcheck

import (
	"net"
	"sync"
)

const ip4Str = "ipv4.ip.test.resource.public.song-zh.com"
const ip6Str = "ipv6.ip.test.resource.public.song-zh.com"

var ip4Once sync.Once
var ip4Ok bool

var ip6Once sync.Once
var ip6Ok bool

func SupportIPv6() bool {
	ip6Once.Do(func() {
		ip6Ok = checkSupportIPv6()
	})
	return ip6Ok
}

func SupportIPv4() bool {
	ip4Once.Do(func() {
		ip4Ok = checkSupportIPv4()
	})
	return ip4Ok
}

func checkSupportIPv4() bool {
	ip4, err := net.ResolveIPAddr("ip4", ip4Str)
	if err != nil {
		return false
	}

	conn, err := net.DialIP("ip4:icmp", nil, ip4) // 使用链路本地多播地址作为目标
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()

	return true
}

func checkSupportIPv6() bool {
	ip6, err := net.ResolveIPAddr("ip6", ip6Str)
	if err != nil {
		return false
	}

	conn, err := net.DialIP("ip6:icmp", nil, ip6) // 使用链路本地多播地址作为目标
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()

	return true
}
