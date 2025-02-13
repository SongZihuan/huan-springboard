package config

type TcpConfig struct {
	RuleList TcpRuleListConfig   `yaml:",inline"`
	Forward  []*TcpForwardConfig `yaml:"forward"`
}

func (t *TcpConfig) setDefault() {
	t.RuleList.setDefault()

	for _, f := range t.Forward {
		f.setDefault()
	}

	return
}

func (t *TcpConfig) check() (err ConfigError) {
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
