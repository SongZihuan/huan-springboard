package database

import (
	"errors"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"gorm.io/gorm"
	"time"
)

func CheckIP(ip string) bool {
	var res BannedIP
	err := db.Model(&BannedIP{}).Where("ip = ?", ip).Order("id desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	} else if err != nil {
		logger.Errorf("CheckIP from DB failed: %s", err.Error())
		return true
	}

	now := time.Now()
	if res.StartAt.Valid && now.Before(res.StartAt.Time) {
		return true // 未生效规则
	} else if res.StopAt.Valid && now.After(res.StopAt.Time) {
		return true // 已失效规则
	}

	return false
}

func CheckLocationNation(nation string) bool {
	if nation == "" {
		return true
	}

	var res BannedLocationNation
	err := db.Model(&BannedLocationNation{}).Where("nation = ?", nation).Order("id desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	} else if err != nil {
		logger.Errorf("CheckLocationNation from DB failed: %s", err.Error())
		return true
	}

	now := time.Now()
	if res.StartAt.Valid && now.Before(res.StartAt.Time) {
		return true // 未生效规则
	} else if res.StopAt.Valid && now.After(res.StopAt.Time) {
		return true // 已失效规则
	}

	return false
}

func CheckLocationProvince(province string) bool {
	if province == "" {
		return true
	}

	var res BannedLocationProvince
	err := db.Model(&BannedLocationProvince{}).Where("province = ?", province).Order("id desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	} else if err != nil {
		logger.Errorf("CheckLocationProvince from DB failed: %s", err.Error())
		return true
	}

	now := time.Now()
	if res.StartAt.Valid && now.Before(res.StartAt.Time) {
		return true // 未生效规则
	} else if res.StopAt.Valid && now.After(res.StopAt.Time) {
		return true // 已失效规则
	}

	return false
}

func CheckLocationCity(city string) bool {
	if city == "" {
		return true
	}

	var res BannedLocationCity
	err := db.Model(&BannedLocationCity{}).Where("city = ?", city).Order("id desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	} else if err != nil {
		logger.Errorf("CheckLocationCity from DB failed: %s", err.Error())
		return true
	}

	now := time.Now()
	if res.StartAt.Valid && now.Before(res.StartAt.Time) {
		return true // 未生效规则
	} else if res.StopAt.Valid && now.After(res.StopAt.Time) {
		return true // 已失效规则
	}

	return false
}

func CheckLocationISP(isp string) bool {
	if isp == "" {
		return true
	}

	var res BannedLocationISP
	err := db.Model(&BannedLocationISP{}).Where("isp = ?", isp).Order("id desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return true
	} else if err != nil {
		logger.Errorf("CheckLocationISP from DB failed: %s", err.Error())
		return true
	}

	now := time.Now()
	if res.StartAt.Valid && now.Before(res.StartAt.Time) {
		return true // 未生效规则
	} else if res.StopAt.Valid && now.After(res.StopAt.Time) {
		return true // 已失效规则
	}

	return false
}
