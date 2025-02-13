package watcher

import (
	"errors"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher

func WatcherConfigFile() error {
	if watcher != nil {
		return nil
	}

	if !config.IsReady() {
		panic("config file path is empty")
	}

	_watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Add a path.
	watcherDir := config.GetConfigFileDir()
	err = _watcher.Add(watcherDir)
	if err != nil {
		return err
	}

	// Start listening for events.
	go func() {
		defer func() {
			err := closeNotifyConfigFile()
			if err != nil {
				logger.Warnf("Auto reload stop with error: %s", err.Error())
				return
			}

			logger.Warnf("Auto reload stop.")
		}()

	OutSideCycle:
		for {
			select {
			case event, ok := <-_watcher.Events:
				if !ok {
					return
				}

				// github.com/fsnotify/fsnotify v1.8.0
				// 根据2024.1月的消息，暂时无法导出RenameFrom，无法跟着重命名
				// issues: https://github.com/fsnotify/fsnotify/issues/630
				if !utils.FilePathEqual(event.Name, config.GetConfigPathFile()) {
					continue OutSideCycle
				}

				if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
					err := config.ReloadConfig()
					if err != nil && err.IsError() {
						logger.Errorf("Config file reload error: %s", err.Error())
					} else if err != nil && err.IsWarning() {
						logger.Warnf("Config file reload error: %s", err.Warning())
					} else {
						logger.Infof("%s", "Config file reload success")
					}
				} else if event.Has(fsnotify.Rename) {
					logger.Warnf("%s", "Config file has been rename")
				} else if event.Has(fsnotify.Remove) {
					logger.Warnf("%s", "Config file has been remove")
				}
			case err, ok := <-_watcher.Errors:
				if !ok || errors.Is(err, fsnotify.ErrClosed) {
					return
				}
				logger.Errorf("Config file notify error: %s", err.Error())
			}
		}
	}()

	watcher = _watcher
	return nil
}

func CloseNotifyConfigFile() {
	_ = closeNotifyConfigFile()
}

func closeNotifyConfigFile() error {
	if watcher == nil {
		return nil
	}

	err := watcher.Close()
	watcher = nil
	return err
}
