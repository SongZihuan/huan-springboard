package database

import (
	"github.com/SongZihuan/huan-springboard/src/config"
	"gorm.io/gorm/logger"
)

// dbLogger 是一个自定义的日志记录器
type dbLogger struct {
	logger.Interface
}

func newDBLogger() *dbLogger {
	if !config.IsReady() {
		panic("config is not ready")
	}

	logLevel := logger.Info
	switch config.GetConfig().LogLevel {
	case "debug":
		fallthrough
	case "info":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	case "error":
		fallthrough
	case "panic":
		logLevel = logger.Error
	case "none":
		logLevel = logger.Silent
	}

	return &dbLogger{
		Interface: logger.Default.LogMode(logLevel),
	}
}

// LogMode 设置日志模式
func (l *dbLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &dbLogger{
		l.Interface.LogMode(level),
	}
}
