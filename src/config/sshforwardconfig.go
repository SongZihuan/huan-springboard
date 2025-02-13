package config

import (
	"fmt"
	"net"
)

type SshForwardConfig struct {
	SrcPoint    int64  `json:"srcpoint"`
	DestAddress string `json:"destaddress"`
}

func (t *SshForwardConfig) setDefault() {
	return
}

func (t *SshForwardConfig) check() (cfgErr ConfigError) {
	if t.SrcPoint < 0 || t.SrcPoint > 65535 {
		return NewConfigError("src point must be between 0 and 65535")
	}

	_, _, err := net.SplitHostPort(t.DestAddress)
	if err != nil {
		return NewConfigError(fmt.Sprintf("dest address not valid: %s", err.Error()))
	}

	return
}
