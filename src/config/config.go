package config

import (
	"github.com/SongZihuan/huan-springboard/src/flagparser"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"sync"
)

type ConfigStruct struct {
	ConfigLock sync.Mutex

	configReady    bool
	yamlHasParser  bool
	sigchan        chan os.Signal
	configPath     string
	configDir      string
	configFileName string
	watcher        *fsnotify.Watcher

	Yaml *YamlConfig
}

func newConfig(configPath string) (*ConfigStruct, error) {
	if configPath == "" {
		if !flagparser.IsReady() {
			panic("flag is not ready")
		}

		configPath = flagparser.ConfigFile()
	}

	configPath, err := utils.CleanFilePathAbs(configPath)
	if err != nil {
		return nil, err
	}

	configDir := filepath.Dir(configPath)
	configFileName := filepath.Base(configPath)

	return &ConfigStruct{
		// Lock不用初始化
		configReady:    false,
		yamlHasParser:  false,
		sigchan:        make(chan os.Signal),
		configPath:     configPath,
		configDir:      configDir,
		configFileName: configFileName,
		Yaml:           nil,
	}, nil
}

func (c *ConfigStruct) Init() (err ConfigError) {
	if c.IsReady() { // 使用IsReady而不是isReady，确保上锁
		return c.Reload()
	}

	initErr := c.init()
	if initErr != nil {
		return NewConfigError("init error: " + initErr.Error())
	}

	parserErr := c.parser(c.configPath)
	if parserErr != nil {
		return NewConfigError("parser error: " + parserErr.Error())
	} else if !c.yamlHasParser {
		return NewConfigError("parser error: unknown")
	}

	c.SetDefault()

	err = c.check()
	if err != nil && err.IsError() {
		return err
	}

	c.configReady = true
	return nil
}

func (c *ConfigStruct) Reload() (err ConfigError) {
	if !c.IsReady() { // 使用IsReady而不是isReady，确保上锁
		return c.Init()
	}

	bak := ConfigStruct{
		configReady:    c.configReady,
		yamlHasParser:  c.yamlHasParser,
		sigchan:        c.sigchan,
		configPath:     c.configPath,
		configDir:      c.configDir,
		configFileName: c.configFileName,
		watcher:        c.watcher,
		Yaml:           c.Yaml,
		// 新建类型
	}

	defer func() {
		if err != nil {
			*c = ConfigStruct{
				configReady:    bak.configReady,
				yamlHasParser:  bak.yamlHasParser,
				sigchan:        bak.sigchan,
				configPath:     bak.configPath,
				configDir:      bak.configDir,
				configFileName: bak.configFileName,
				watcher:        bak.watcher,
				Yaml:           bak.Yaml,
				// 新建类型 Lock不需要复制
			}
		}
	}()

	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()

	reloadErr := c.reload()
	if reloadErr != nil {
		return NewConfigError("reload error: " + reloadErr.Error())
	}

	parserErr := c.parser(c.configPath)
	if parserErr != nil {
		return NewConfigError("reload parser error: " + parserErr.Error())
	} else if !c.yamlHasParser {
		return NewConfigError("reload parser error: unknown")
	}

	c.SetDefault()

	err = c.check()
	if err != nil && err.IsError() {
		return err
	}

	c.configReady = true
	return nil
}

func (c *ConfigStruct) clear() error {
	c.configReady = false
	c.yamlHasParser = false
	// sigchan和watcher 不变
	c.Yaml = nil
	return nil
}

func (c *ConfigStruct) parser(filepath string) ParserError {
	err := c.Yaml.parser(filepath)
	if err != nil {
		return err
	}

	c.yamlHasParser = true
	return nil
}

func (c *ConfigStruct) SetDefault() {
	if !c.yamlHasParser {
		panic("yaml must parser first")
	}

	c.Yaml.setDefault()
}

func (c *ConfigStruct) check() (err ConfigError) {
	err = c.Yaml.check()
	if err != nil && err.IsError() {
		return err
	}

	return nil
}

func (c *ConfigStruct) isReady() bool {
	return c.yamlHasParser && c.configReady
}

func (c *ConfigStruct) init() error {
	c.configReady = false
	c.yamlHasParser = false

	err := initSignal(c.sigchan)
	if err != nil {
		return err
	}

	c.Yaml = new(YamlConfig)
	err = c.Yaml.Init()
	if err != nil {
		return err
	}

	return nil
}

func (c *ConfigStruct) reload() error {
	err := c.clear()
	if err != nil {
		return err
	}

	c.Yaml = new(YamlConfig)
	err = c.Yaml.Init()
	if err != nil {
		return err
	}

	return nil
}

// export func

func (c *ConfigStruct) IsReady() bool {
	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()
	return c.isReady()
}

func (c *ConfigStruct) GetSignalChan() chan os.Signal {
	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()

	return c.sigchan
}

func (c *ConfigStruct) GetConfig() *YamlConfig {
	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()

	if !c.isReady() {
		panic("config is not ready")
	}

	return c.Yaml
}

func (c *ConfigStruct) GetConfigPathFile() string {
	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()

	// 不需要检查Ready

	return c.configPath
}

func (c *ConfigStruct) GetConfigFileDir() string {
	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()

	// 不需要检查Ready

	return c.configDir
}

func (c *ConfigStruct) GetConfigFileName() string {
	c.ConfigLock.Lock()
	defer c.ConfigLock.Unlock()

	// 不需要检查Ready
	return c.configFileName
}
