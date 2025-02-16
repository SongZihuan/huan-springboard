package config

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/ipcheck"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"net"
)

type SshForwardConfig struct {
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

	CountRules []*SshCountRuleConfig `yaml:"count-rules"` // 全局连接规则

	ResolveIPv4SrcAddress  *net.TCPAddr `yaml:"-"`
	ResolveIPv4DestAddress *net.TCPAddr `yaml:"-"`

	ResolveIPv6SrcAddress  *net.TCPAddr `yaml:"-"`
	ResolveIPv6DestAddress *net.TCPAddr `yaml:"-"`

	Cross bool `yaml:"-"` // 开启交叉
}

func (s *SshForwardConfig) setDefault() {
	if s.SrcPort == 0 {
		s.SrcPort = 22
	}

	s.AllowCross.SetDefaultEnable()

	s.IPv4SrcServerProxy.SetDefaultEnable()
	s.IPv6SrcServerProxy.SetDefaultEnable()

	s.IPv4DestRequestProxy.SetDefaultDisable()
	s.IPv6DestRequestProxy.SetDefaultDisable()

	if s.IPv4DestRequestProxyVersion <= 0 && s.IPv4DestRequestProxyVersion != -1 { // -1 表示使用最新版; 0 表示默认（使用版本1）
		s.IPv4DestRequestProxyVersion = 1
	}

	if s.IPv6DestRequestProxyVersion <= 0 && s.IPv6DestRequestProxyVersion != -1 { // -1 表示使用最新版; 0 表示默认（使用版本1）
		s.IPv6DestRequestProxyVersion = 1
	}

	for _, r := range s.CountRules {
		r.setDefault()
	}

	return
}

func (s *SshForwardConfig) check() (cfgErr ConfigError) {
	if s.SrcPort <= 0 || s.SrcPort > 65535 { // 一般不建议使用端口号0
		return NewConfigError("src point must be between 1 and 65535")
	}

	if s.IPv4DestRequestProxy.IsEnable(false) || s.IPv6DestRequestProxy.IsEnable(false) {
		_ = NewConfigWarning("ssh does not recommend using proxy protocol")
	}

	if ipcheck.SupportIPv4() {
		if s.IPv4DestAddress != "" {
			ip4, err := net.ResolveTCPAddr("tcp4", s.IPv4DestAddress)
			if err != nil {
				return NewConfigError(fmt.Sprintf("ipv4 dest address not valid: %s", err.Error()))
			}

			s.ResolveIPv4DestAddress = ip4
		} else if s.DestAddress != "" {
			ip4, err := net.ResolveTCPAddr("tcp4", s.DestAddress)
			if err == nil {
				s.ResolveIPv4DestAddress = ip4
			}
		} else if s.AllowCross.IsEnable() && s.IPv6DestAddress != "" {
			// 如果 IPv6DestAddress 可以解析为 ipv4 那么就可以直接转发
			ip4, err := net.ResolveTCPAddr("tcp4", s.IPv6DestAddress)
			if err == nil {
				s.ResolveIPv4DestAddress = ip4
			}
		}
	}

	if ipcheck.SupportIPv6() {
		if s.IPv6DestAddress != "" {
			ip6, err := net.ResolveTCPAddr("tcp6", s.IPv6DestAddress)
			if err != nil {
				return NewConfigError(fmt.Sprintf("ipv6 dest address not valid: %s", err.Error()))
			}

			s.ResolveIPv6DestAddress = ip6
		} else if s.DestAddress != "" {
			ip6, err := net.ResolveTCPAddr("tcp6", s.DestAddress)
			if err == nil {
				s.ResolveIPv6DestAddress = ip6
			}
		} else if s.AllowCross.IsEnable() && s.IPv4DestAddress != "" {
			// 如果 IPv4DestAddress 可以解析为 ipv6 那么就可以直接转发
			ip6, err := net.ResolveTCPAddr("tcp6", s.IPv4DestAddress)
			if err == nil {
				s.ResolveIPv6DestAddress = ip6
			}
		}
	}

	{
		ip4, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", s.SrcPort))
		if err != nil {
			return NewConfigError(fmt.Sprintf("ipv4 src address not valid: %s", err.Error()))
		}

		s.ResolveIPv4SrcAddress = ip4

		ip6, err := net.ResolveTCPAddr("tcp6", fmt.Sprintf(":%d", s.SrcPort))
		if err != nil {
			return NewConfigError(fmt.Sprintf("ipv6 src address not valid: %s", err.Error()))
		}

		s.ResolveIPv6SrcAddress = ip6
	}

	if s.ResolveIPv4DestAddress == nil && s.ResolveIPv6DestAddress == nil {
		return NewConfigError("dest address not valid")
	}

	s.Cross = s.AllowCross.IsEnable(true) && ipcheck.SupportIPv4() && ipcheck.SupportIPv6() && (s.ResolveIPv4DestAddress == nil || s.ResolveIPv6DestAddress == nil)

	tr := int64(-1)
	ms := int64(-1)
	for _, r := range s.CountRules {
		err := r.check()
		if err != nil && err.IsError() {
			return err
		}

		if (tr != -1 && ms != -1) && r.TryCount > tr {
			return NewConfigError("The count-rules are not sorted correctly, the try-count with the largest number is placed first")
		} else if (tr != -1 && ms != -1) && r.Seconds > ms {
			return NewConfigError("The count-rules are not sorted correctly, the seconds with the largest number is placed first")
		} else {
			tr = r.TryCount
			ms = r.Seconds
		}
	}

	return nil
}
