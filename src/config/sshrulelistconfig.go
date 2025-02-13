package config

import "github.com/SongZihuan/huan-springboard/src/utils"

type SshRuleListConfig struct {
	RuleList                []*SshRuleConfig `yaml:"rules"`
	DefaultBanned           utils.StringBool `yaml:"default-banned"`
	AlwaysAllIntranet       utils.StringBool `yaml:"always-all-intranet"`
	StatisticalPeriodSecond uint64           `yaml:"statistical-period-second"`
	ConnectCountOfCycle     uint64           `yaml:"connect-count-of-cycle"`
}

func (t *SshRuleListConfig) setDefault() {
	for _, r := range t.RuleList {
		r.setDefault()
	}

	t.DefaultBanned.SetDefaultDisable()
	t.AlwaysAllIntranet.SetDefaultEnable()

	if t.StatisticalPeriodSecond <= 0 {
		t.StatisticalPeriodSecond = 5
	}

	if t.ConnectCountOfCycle <= 0 {
		t.ConnectCountOfCycle = 0
	}

	return
}

func (t *SshRuleListConfig) check() (err ConfigError) {
	if len(t.RuleList) == 0 {
		return nil
	}

	for _, r := range t.RuleList {
		err := r.check()
		if err != nil && err.IsError() {
			return err
		}
	}

	return nil
}
