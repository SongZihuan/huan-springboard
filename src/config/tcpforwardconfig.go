package config

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"net"
)

type TcpForwardConfig struct {
	SrcPort                 int64            `yaml:"src"`
	DestAddress             string           `yaml:"dest"`
	SrcServerProxy          utils.StringBool `yaml:"src-proxy"`
	DestRequestProxy        utils.StringBool `yaml:"dest-proxy"`
	DestRequestProxyVersion int              `yaml:"dest-proxy-version"`
}

func (t *TcpForwardConfig) setDefault() {
	t.SrcServerProxy.SetDefaultEnable()
	t.DestRequestProxy.SetDefaultEnable()

	if t.DestRequestProxyVersion <= 0 && t.DestRequestProxyVersion != -1 { // -1 表示使用最新版; 0 表示默认（使用版本2）
		t.DestRequestProxyVersion = 2
	}

	if t.DestRequestProxyVersion == 1 {
		_ = NewConfigError("TCP proxy protocol version 2 (or higher) is recommended.")
	}

	return
}

func (t *TcpForwardConfig) check() (cfgErr ConfigError) {
	if t.SrcPort < 0 || t.SrcPort > 65535 {
		return NewConfigError("src point must be between 0 and 65535")
	}

	_, _, err := net.SplitHostPort(t.DestAddress)
	if err != nil {
		return NewConfigError(fmt.Sprintf("dest address not valid: %s", err.Error()))
	}

	return
}
