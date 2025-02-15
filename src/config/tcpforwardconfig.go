package config

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/ipcheck"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"net"
)

type TcpForwardConfig struct {
	SrcPort         int64            `yaml:"src"`
	DestAddress     string           `yaml:"dest"`
	IPv4DestAddress string           `yaml:"ipv4-dest"`
	IPv6DestAddress string           `yaml:"ipv6-dest"`
	AllowCross      utils.StringBool `yaml:"allow-cross"` // 允许 ipv4 -> ipv6 或 ipv6 -> ipv4

	IPv4SrcServerProxy utils.StringBool `yaml:"ipv4-src-proxy"`
	IPv6SrcServerProxy utils.StringBool `yaml:"ipv6-src-proxy"`

	IPv4DestRequestProxy        utils.StringBool `yaml:"ipv4-dest-proxy"`
	IPv4DestRequestProxyVersion int              `yaml:"ipv4-dest-proxy-version"`

	IPv6DestRequestProxy        utils.StringBool `yaml:"ipv6-dest-proxy"`
	IPv6DestRequestProxyVersion int              `yaml:"ipv6-dest-proxy-version"`

	ResolveIPv4SrcAddress  *net.TCPAddr `yaml:"-"`
	ResolveIPv4DestAddress *net.TCPAddr `yaml:"-"`

	ResolveIPv6SrcAddress  *net.TCPAddr `yaml:"-"`
	ResolveIPv6DestAddress *net.TCPAddr `yaml:"-"`

	Cross bool `yaml:"-"` // 开启交叉
}

func (t *TcpForwardConfig) setDefault() {
	t.AllowCross.SetDefaultEnable()

	t.IPv4SrcServerProxy.SetDefaultEnable()
	t.IPv6SrcServerProxy.SetDefaultEnable()

	t.IPv4DestRequestProxy.SetDefaultEnable()
	t.IPv6DestRequestProxy.SetDefaultEnable()

	if t.IPv4DestRequestProxyVersion <= 0 && t.IPv4DestRequestProxyVersion != -1 { // -1 表示使用最新版; 0 表示默认（使用版本1）
		t.IPv4DestRequestProxyVersion = 1
	}

	if t.IPv6DestRequestProxyVersion <= 0 && t.IPv6DestRequestProxyVersion != -1 { // -1 表示使用最新版; 0 表示默认（使用版本1）
		t.IPv6DestRequestProxyVersion = 1
	}

	return
}

func (t *TcpForwardConfig) check() (cfgErr ConfigError) {
	if t.SrcPort <= 0 || t.SrcPort > 65535 { // 一般不建议使用端口号0
		return NewConfigError("src point must be between 1 and 65535")
	}

	if ipcheck.SupportIPv4() {
		if t.IPv4DestAddress != "" {
			ip4, err := net.ResolveTCPAddr("tcp4", t.IPv4DestAddress)
			if err != nil {
				return NewConfigError(fmt.Sprintf("ipv4 dest address not valid: %s", err.Error()))
			}

			t.ResolveIPv4DestAddress = ip4
		} else if t.DestAddress != "" {
			ip4, err := net.ResolveTCPAddr("tcp4", t.DestAddress)
			if err == nil {
				t.ResolveIPv4DestAddress = ip4
			}
		} else if t.AllowCross.IsEnable() && t.IPv6DestAddress != "" {
			// 如果 IPv6DestAddress 可以解析为 ipv4 那么就可以直接转发
			ip4, err := net.ResolveTCPAddr("tcp4", t.IPv6DestAddress)
			if err == nil {
				t.ResolveIPv4DestAddress = ip4
			}
		}
	}

	if ipcheck.SupportIPv6() {
		if t.IPv6DestAddress != "" {
			ip6, err := net.ResolveTCPAddr("tcp6", t.IPv6DestAddress)
			if err != nil {
				return NewConfigError(fmt.Sprintf("ipv6 dest address not valid: %s", err.Error()))
			}

			t.ResolveIPv6DestAddress = ip6
		} else if t.DestAddress != "" {
			ip6, err := net.ResolveTCPAddr("tcp6", t.DestAddress)
			if err == nil {
				t.ResolveIPv6DestAddress = ip6
			}
		} else if t.AllowCross.IsEnable() && t.IPv4DestAddress != "" {
			// 如果 IPv4DestAddress 可以解析为 ipv6 那么就可以直接转发
			ip6, err := net.ResolveTCPAddr("tcp6", t.IPv4DestAddress)
			if err == nil {
				t.ResolveIPv6DestAddress = ip6
			}
		}
	}

	{
		ip4, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", t.SrcPort))
		if err != nil {
			return NewConfigError(fmt.Sprintf("ipv4 src address not valid: %s", err.Error()))
		}

		t.ResolveIPv4SrcAddress = ip4

		ip6, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf(":%d", t.SrcPort))
		if err != nil {
			return NewConfigError(fmt.Sprintf("ipv6 src address not valid: %s", err.Error()))
		}

		t.ResolveIPv6SrcAddress = ip6
	}

	if t.ResolveIPv4DestAddress == nil && t.ResolveIPv6DestAddress == nil {
		return NewConfigError("dest address not valid")
	}

	t.Cross = t.AllowCross.IsEnable(true) && ipcheck.SupportIPv4() && ipcheck.SupportIPv6() && (t.ResolveIPv4DestAddress == nil || t.ResolveIPv6DestAddress == nil)

	return
}
