package config

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/flagparser"
	"gopkg.in/yaml.v3"
	"os"
)

func InitConfig(configPath string) ConfigError {
	var err error

	config, err = newConfig(configPath)
	if err != nil {
		return NewConfigError(err.Error())
	}

	cfgErr := config.Init()
	if cfgErr != nil && cfgErr.IsError() {
		return cfgErr
	}

	if !config.IsReady() {
		return NewConfigError("config not ready")
	}

	err = OutputConfig()
	if err != nil {
		fmt.Printf("output config error: %s\n", err.Error())
	}

	return nil
}

func ReloadConfig() ConfigError {
	cfgErr := config.Reload()
	if cfgErr != nil && cfgErr.IsError() {
		return cfgErr
	}

	if !config.IsReady() {
		return NewConfigError("config not ready")
	}

	err := OutputConfig()
	if err != nil {
		fmt.Printf("output config error: %s\n", err.Error())
	}

	return nil
}

func OutputConfig() error {
	outputPath := flagparser.OutputConfigFile()
	if outputPath == "" {
		return nil
	}

	out, err := yaml.Marshal(config.Yaml)
	if err != nil {
		return err
	}

	err = os.WriteFile(outputPath, out, 0644)
	if err != nil {
		return err
	}

	return nil
}

func IsReady() bool {
	return config.IsReady()
}

func GetConfig() *YamlConfig {
	return config.GetConfig()
}

func GetSignalChan() chan os.Signal {
	return config.GetSignalChan()
}

func GetConfigPathFile() string {
	return config.GetConfigPathFile()
}

func GetConfigFileDir() string {
	return config.GetConfigFileDir()
}

func GetConfigFileName() string {
	return config.GetConfigFileName()
}

var config *ConfigStruct
