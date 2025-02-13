package config

type TcpRuleConfig struct {
	RuleConfig `yaml:",inline"`
}

func (t *TcpRuleConfig) setDefault() {
	t.RuleConfig.setDefault()
	return
}

func (t *TcpRuleConfig) check() (err ConfigError) {
	err = t.RuleConfig.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}
