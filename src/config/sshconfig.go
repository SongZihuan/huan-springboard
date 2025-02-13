package config

type SshConfig struct {
	RuleList SshRuleListConfig   `yaml:",inline"`
	Forward  []*SshForwardConfig `yaml:"forward"`
}

func (t *SshConfig) setDefault() {
	t.RuleList.setDefault()

	for _, f := range t.Forward {
		f.setDefault()
	}

	return
}

func (t *SshConfig) check() (err ConfigError) {
	err = t.RuleList.check()
	if err != nil && err.IsError() {
		return err
	}

	for _, f := range t.Forward {
		err = f.check()
		if err != nil && err.IsError() {
			return err
		}
	}

	return
}
