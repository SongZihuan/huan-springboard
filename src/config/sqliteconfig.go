package config

import "github.com/SongZihuan/huan-springboard/src/utils"

type SQLiteConfig struct {
	Path        string           `yaml:"path"`
	ActiveClose utils.StringBool `yaml:"active-close"`
	Clean       DBCleanConfig    `yaml:"clean"`
}

func (s *SQLiteConfig) setDefault() {
	s.ActiveClose.SetDefaultDisable()
	s.Clean.setDefault()
	return
}

func (s *SQLiteConfig) check() (err ConfigError) {
	if s.Path == "" {
		return NewConfigError("sqlite path is empty")
	}

	err = s.Clean.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}
