package config

import (
	"github.com/SongZihuan/huan-springboard/src/iface"
	"github.com/SongZihuan/huan-springboard/src/utils"
)

type TcpRuleListConfig struct {
	RuleList                []*TcpRuleConfig `yaml:"rules"`
	DefaultBanned           utils.StringBool `yaml:"default-banned"`
	AlwaysAllIntranet       utils.StringBool `yaml:"always-all-intranet"`
	InterfaceName           string           `yaml:"interface-name"`
	StatisticalPeriodSecond uint64           `yaml:"statistical-period-second"`
	ReceiveBytesOfCycle     uint64           `yaml:"receive-bytes-of-cycle"`
	TransmitBytesOfCycle    uint64           `yaml:"transmit-bytes-of-cycle"`
}

func (t *TcpRuleListConfig) setDefault() {
	for _, r := range t.RuleList {
		r.setDefault()
	}

	t.DefaultBanned.SetDefaultDisable()
	t.AlwaysAllIntranet.SetDefaultEnable()

	if t.InterfaceName != "" {
		if t.StatisticalPeriodSecond <= 0 {
			t.StatisticalPeriodSecond = 0
		}

		if t.ReceiveBytesOfCycle <= 0 {
			t.ReceiveBytesOfCycle = 0
		}

		if t.TransmitBytesOfCycle <= 0 {
			t.TransmitBytesOfCycle = 0
		}
	}

	return
}

func (t *TcpRuleListConfig) check() (err ConfigError) {
	if len(t.RuleList) == 0 {
		return nil
	}

	for _, r := range t.RuleList {
		err := r.check()
		if err != nil && err.IsError() {
			return err
		}
	}

	if t.InterfaceName != "" {
		if _, ok := iface.Iface[t.InterfaceName]; !ok {
			return NewConfigError("bad interface name")
		}
	}

	return nil
}
