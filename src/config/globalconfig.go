package config

import (
	"fmt"
	resource "github.com/SongZihuan/huan-springboard"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"os"
)

const EnvModeName = "HUAN_SPRINGBOARD_MODE"

const (
	DebugMode   = "debug"
	ReleaseMode = "release"
	TestMode    = "test"
)

type LoggerLevel string

var levelMap = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
	"panic": true,
	"none":  true,
}

type GlobalConfig struct {
	Mode     string           `yaml:"mode"`
	LogLevel string           `yaml:"log-level"`
	LogTag   utils.StringBool `yaml:"log-tag"`
	Timezone string           `yaml:"time-zone"`
	Name     string           `yaml:"name"`
	Quite    utils.StringBool `yaml:"quite"`

	SystemName string `yaml:"-"`
}

func (g *GlobalConfig) setDefault() {
	if g.Mode == "" {
		g.Mode = os.Getenv(EnvModeName)
	}

	if g.Mode == "" {
		g.Mode = DebugMode
	}

	_ = os.Setenv(EnvModeName, g.Mode)

	if g.LogLevel == "" && (g.Mode == DebugMode || g.Mode == TestMode) {
		g.LogLevel = "debug"
	} else if g.LogLevel == "" {
		g.LogLevel = "warn"
	}

	if g.Mode == DebugMode || g.Mode == TestMode {
		g.LogTag.SetDefaultEnable()
	} else {
		g.LogTag.SetDefaultDisable()
	}

	if g.Timezone == "" {
		g.Timezone = "Local"
	}

	if g.Name == "" {
		g.Name = "dev"
	}

	g.Quite.SetDefaultDisable()

	return
}

func (g *GlobalConfig) Check() ConfigError {
	if g.Mode != DebugMode && g.Mode != ReleaseMode && g.Mode != TestMode {
		return NewConfigError("bad mode")
	}

	if _, ok := levelMap[g.LogLevel]; !ok {
		return NewConfigError("log level error")
	}

	g.SystemName = fmt.Sprintf("%s-%s", resource.Name, g.Name)

	return nil
}

func (g *GlobalConfig) GetRunMode() string {
	return g.Mode
}

func (g *GlobalConfig) IsDebug() bool {
	return g.Mode == DebugMode
}

func (g *GlobalConfig) IsRelease() bool {
	return g.Mode == ReleaseMode
}

func (g *GlobalConfig) IsTest() bool {
	return g.Mode == TestMode
}
