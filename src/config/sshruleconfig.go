package config

type SshRuleConfig struct {
	RuleConfig `yaml:",inline"`
}

func (s *SshRuleConfig) setDefault() {
	s.RuleConfig.setDefault()

	return
}

func (s *SshRuleConfig) check() (err ConfigError) {
	err = s.RuleConfig.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}
