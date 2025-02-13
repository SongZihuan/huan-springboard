package config

import "github.com/SongZihuan/huan-springboard/src/utils"

type SQLiteConfig struct {
	Path        string           `yaml:"path"`
	ActiveClose utils.StringBool `yaml:"active-close"`
}

func (s *SQLiteConfig) setDefault() {
	s.ActiveClose.SetDefaultDisable()
	return
}

func (s *SQLiteConfig) check() (cfgErr ConfigError) {
	if s.Path == "" {
		return NewConfigError("sqlite path is empty")
	}
	return nil
}
