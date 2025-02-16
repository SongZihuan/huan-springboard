package config

import "github.com/SongZihuan/huan-springboard/src/utils"

type SshRuleListConfig struct {
	RuleList   []*SshRuleConfig      `yaml:"rules"`
	CountRules []*SshCountRuleConfig `yaml:"count-rules"` // 全局连接规则

	DefaultBanned       utils.StringBool `yaml:"default-banned"`        // 默认（未名字规则）拒绝连接
	AlwaysAllowIntranet utils.StringBool `yaml:"always-allow-intranet"` // 总是允许内网连接（配置 ip 数据库封禁除外）
	AlwaysAllowLoopback utils.StringBool `yaml:"always-allow-loopback"` // 总是允许本地回环地址连接（不检查 ip 数据库封禁）
}

func (s *SshRuleListConfig) setDefault() {
	for _, r := range s.RuleList {
		r.setDefault()
	}

	for _, r := range s.CountRules {
		r.setDefault()
	}

	s.DefaultBanned.SetDefaultEnable()
	s.AlwaysAllowIntranet.SetDefaultDisable()
	s.AlwaysAllowLoopback.SetDefaultEnable()

	return
}

func (s *SshRuleListConfig) check() (err ConfigError) {
	if !s.DefaultBanned.IsEnable(false) {
		_ = NewConfigWarning("ssh recommends setting the default policy to banned")
	}

	for _, r := range s.RuleList {
		err := r.check()
		if err != nil && err.IsError() {
			return err
		}
	}

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
