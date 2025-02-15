package database

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitSQLite() error {
	if !config.IsReady() {
		panic("config is not ready")
	}

	_db, err := gorm.Open(sqlite.Open(config.GetConfig().SQLite.Path), &gorm.Config{
		Logger: newDBLogger(),
	})
	if err != nil {
		return fmt.Errorf("connect to sqlite (%s) failed: %s", config.GetConfig().SQLite.Path, err)
	}

	err = _db.AutoMigrate(&TcpBannedIP{}, &TcpBannedLocationNation{},
		&TcpBannedLocationProvince{}, &TcpBannedLocationCity{},
		&TcpBannedLocationISP{}, &SshBannedIP{}, &SshBannedLocationNation{},
		&SshBannedLocationProvince{}, &SshBannedLocationCity{},
		&SshBannedLocationISP{}, &SshConnectRecord{}, &IfaceRecord{})
	if err != nil {
		return fmt.Errorf("auto migrate sqlite (%s) failed: %s", config.GetConfig().SQLite.Path, err)
	}

	db = _db
	return nil
}

func CloseSQLite() {
	if db == nil {
		return
	}

	defer func() {
		db = nil
	}()

	if config.GetConfig().SQLite.ActiveClose.IsEnable(false) {
		// https://github.com/go-gorm/gorm/issues/3145
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}
