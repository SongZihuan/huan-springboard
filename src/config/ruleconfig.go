package config

import (
	"github.com/SongZihuan/huan-springboard/src/utils"
	"net"
)

type RuleType string

type RuleConfig struct {
	Nation        string           `yaml:"nation"`
	NationVague   string           `yaml:"nation-vague"`
	Province      string           `yaml:"province"`
	ProvinceVague string           `yaml:"province-vague"`
	City          string           `yaml:"city"`
	CityVague     string           `yaml:"city-vague"`
	ISP           string           `yaml:"isp"`
	ISPVague      string           `yaml:"isp-vague"`
	IPv4          string           `yaml:"ipv4"`
	IPv6          string           `yaml:"ipv6"`
	IPv4Cidr      string           `yaml:"ipv4cidr"`
	IPv6Cidr      string           `yaml:"ipv6cidr"`
	Banned        utils.StringBool `yaml:"banned"`
}

func (r *RuleConfig) setDefault() {
	r.Banned.SetDefaultEnable()
	return
}

func (r *RuleConfig) check() (err ConfigError) {
	if r.IPv4 != "" {
		if !utils.IsValidIPv4(r.IPv4) {
			return NewConfigError("bad IPv4")
		}
	}

	if r.IPv6 != "" {
		if !utils.IsValidIPv6(r.IPv6) {
			return NewConfigError("bad IPv6")
		}
	}

	if r.IPv4Cidr != "" {
		if !utils.IsValidIPv4CIDR(r.IPv4Cidr) {
			return NewConfigError("bad IPv4")
		}
	}

	if r.IPv6Cidr != "" {
		if !utils.IsValidIPv6CIDR(r.IPv6Cidr) {
			return NewConfigError("bad IPv6")
		}
	}

	if r.IPv4 == "" && r.IPv6 == "" && r.IPv4Cidr == "" && r.IPv6Cidr == "" {
		return NewConfigError("bad IP or CIDR")
	}

	return nil
}

func (r *RuleConfig) HasLocation() bool {
	return r.Nation != "" || r.NationVague != "" ||
		r.Province != "" || r.ProvinceVague != "" ||
		r.City != "" || r.CityVague != "" ||
		r.ISP != "" || r.ISPVague != ""
}

func (r *RuleConfig) CheckIP(ip net.IP) (bool, error) {
	if r.IPv4 != "" {
		ip4 := net.ParseIP(r.IPv4)
		if ip4 != nil && ip4.Equal(ip) {
			return true, nil
		}
	}

	if r.IPv6 != "" {
		ip6 := net.ParseIP(r.IPv6)
		if ip6 != nil && ip6.Equal(ip) {
			return true, nil
		}
	}

	if r.IPv4Cidr != "" {
		_, ipnet, err := net.ParseCIDR(r.IPv4Cidr)
		if err == nil && ipnet.Contains(ip) {
			return true, nil
		}
	}

	if r.IPv6Cidr != "" {
		_, ipnet, err := net.ParseCIDR(r.IPv6Cidr)
		if err == nil && ipnet.Contains(ip) {
			return true, nil
		}
	}

	if r.IPv4Cidr == "0.0.0.0/0" {
		return true, nil
	}

	return false, nil
}
