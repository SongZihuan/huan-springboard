package config

type ApiConfig struct {
	AppCode string `yaml:"app-code"`
}

func (a *ApiConfig) setDefault() {
	return
}

func (a *ApiConfig) check() (err ConfigError) {
	if a.AppCode == "" {
		return NewConfigError("app-code is empty")
	}

	return nil
}
