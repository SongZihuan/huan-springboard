package config

import "net"

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func (r *RedisConfig) setDefault() {
	if r.DB <= 0 {
		r.DB = 0
	}
	return
}

func (r *RedisConfig) check() (cfgErr ConfigError) {
	_, _, err := net.SplitHostPort(r.Address)
	if err != nil {
		return NewConfigError("redis address is invalid")
	}

	return nil
}
