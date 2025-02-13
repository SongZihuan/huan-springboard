package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type YamlConfig struct {
	GlobalConfig `yaml:",inline"`

	TCP    TcpConfig    `yaml:"tcp"`
	SSH    SshConfig    `yaml:"ssh"`
	API    ApiConfig    `yaml:"api"`
	Redis  RedisConfig  `yaml:"redis"`
	SQLite SQLiteConfig `yaml:"sqlite"`
}

func (y *YamlConfig) Init() error {
	return nil
}

func (y *YamlConfig) setDefault() {
	y.GlobalConfig.setDefault()
	y.TCP.setDefault()
	y.SSH.setDefault()
	y.API.setDefault()
	y.Redis.setDefault()
	y.SQLite.setDefault()
}

func (y *YamlConfig) check() (err ConfigError) {
	err = y.GlobalConfig.Check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.TCP.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.SSH.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.API.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.Redis.check()
	if err != nil && err.IsError() {
		return err
	}

	err = y.SQLite.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}

func (y *YamlConfig) parser(filepath string) ParserError {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return NewParserError(err, err.Error())
	}

	err = yaml.Unmarshal(file, y)
	if err != nil {
		return NewParserError(err, err.Error())
	}

	return nil
}
