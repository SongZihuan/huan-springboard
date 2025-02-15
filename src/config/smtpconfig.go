package config

type SMTPConfig struct {
	Address   string   `yaml:"address"`
	User      string   `yaml:"user"`
	Password  string   `yaml:"password"`
	Recipient []string `yaml:"recipient"`
}

func (s *SMTPConfig) setDefault() {
	return
}

func (s *SMTPConfig) check() (err ConfigError) {
	return nil
}
