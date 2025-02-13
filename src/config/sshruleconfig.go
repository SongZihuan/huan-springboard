package config

type SshRuleConfig struct {
	BaseRule RuleConfig `yaml:",inline"`
}

func (t *SshRuleConfig) setDefault() {
	t.BaseRule.setDefault()
	return
}

func (t *SshRuleConfig) check() (err ConfigError) {
	err = t.BaseRule.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}
