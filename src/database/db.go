package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"gorm.io/gorm"
	"net"
	"strings"
	"time"
)

var ErrNotFound = fmt.Errorf("not found")

func TcpCheckIP(ip string) bool {
	var res TcpBannedIP
	err := db.Model(&TcpBannedIP{}).Where("ip = ?", ip).Order("id desc").First(&res).Error
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

func TcpCheckLocationNation(nation string) bool {
	if nation == "" {
		return true
	}

	var res TcpBannedLocationNation
	err := db.Model(&TcpBannedLocationNation{}).Where("nation = ?", nation).Order("id desc").First(&res).Error
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

func TcpCheckLocationProvince(province string) bool {
	if province == "" {
		return true
	}

	var res TcpBannedLocationProvince
	err := db.Model(&TcpBannedLocationProvince{}).Where("province = ?", province).Order("id desc").First(&res).Error
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

func TcpCheckLocationCity(city string) bool {
	if city == "" {
		return true
	}

	var res TcpBannedLocationCity
	err := db.Model(&TcpBannedLocationCity{}).Where("city = ?", city).Order("id desc").First(&res).Error
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

func TcpCheckLocationISP(isp string) bool {
	if isp == "" {
		return true
	}

	var res TcpBannedLocationISP
	err := db.Model(&TcpBannedLocationISP{}).Where("isp = ?", isp).Order("id desc").First(&res).Error
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

func SshCheckIP(ip string) bool {
	var res SshBannedIP
	err := db.Model(&SshBannedIP{}).Where("ip = ?", ip).Order("id desc").First(&res).Error
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

func SshCheckLocationNation(nation string) bool {
	if nation == "" {
		return true
	}

	var res SshBannedLocationNation
	err := db.Model(&SshBannedLocationNation{}).Where("nation = ?", nation).Order("id desc").First(&res).Error
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

func SshCheckLocationProvince(province string) bool {
	if province == "" {
		return true
	}

	var res SshBannedLocationProvince
	err := db.Model(&SshBannedLocationProvince{}).Where("province = ?", province).Order("id desc").First(&res).Error
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

func SshCheckLocationCity(city string) bool {
	if city == "" {
		return true
	}

	var res SshBannedLocationCity
	err := db.Model(&SshBannedLocationCity{}).Where("city = ?", city).Order("id desc").First(&res).Error
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

func SshCheckLocationISP(isp string) bool {
	if isp == "" {
		return true
	}

	var res SshBannedLocationISP
	err := db.Model(&SshBannedLocationISP{}).Where("isp = ?", isp).Order("id desc").First(&res).Error
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

func AddSshConnectRecord(from string, fromIP net.IP, to *net.TCPAddr, accept bool, t time.Time, mark string) (*SshConnectRecord, error) {
	if fromIP == nil {
		fromIP = net.ParseIP(from)
		if fromIP == nil {
			ipStr, _, err := net.SplitHostPort(from)
			if err != nil {
				return nil, err
			}

			fromIP = net.ParseIP(ipStr)
			if fromIP == nil {
				return nil, fmt.Errorf("from is not ip (or with port)")
			}
		}
	}

	if mark != "" && !strings.HasSuffix(mark, "。") {
		mark += "。"
	}

	record := SshConnectRecord{
		From:   fromIP.String(),
		To:     to.String(),
		Accept: accept,
		Time:   t,
		Mark:   mark,
	}
	err := db.Create(&record).Error
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func UpdateSshConnectRecord(record *SshConnectRecord, mark string) (err error) {
	defer func() {
		// 有除法，防止零除
		r := recover()
		if r != nil && err == nil {
			if _err, ok := r.(error); ok {
				err = _err
			} else {
				err = fmt.Errorf("%v", err)
			}
		}
	}()

	if record == nil {
		return fmt.Errorf("record is nil")
	}

	if mark != "" && !strings.HasSuffix(mark, "。") {
		mark += "。"
	}

	record.TimeConsuming = sql.NullInt64{
		Valid: true,
		Int64: int64(time.Since(record.Time) / time.Millisecond),
	}

	record.Mark = record.Mark + mark

	err = db.Save(record).Error // record已经是指针
	if err != nil {
		return err
	}

	return nil
}

func FindSshConnectRecord(from string, fromIP net.IP, to *net.TCPAddr, limit int, after time.Time) ([]SshConnectRecord, error) {
	var res []SshConnectRecord

	if fromIP == nil {
		fromIP = net.ParseIP(from)
		if fromIP == nil {
			ipStr, _, err := net.SplitHostPort(from)
			if err != nil {
				return nil, err
			}

			fromIP = net.ParseIP(ipStr)
			if fromIP == nil {
				return nil, fmt.Errorf("from is not ip (or with port)")
			}
		}
	}

	err := db.Model(&SshConnectRecord{}).Where("`time` > ? AND `to` = ? AND `from` = ?", after, to.String(), fromIP.String()).Order("time asc").Limit(limit).Find(&res).Error
	if err != nil {
		return nil, err
	}

	return res, nil
}

func AddIfaceRecord(name string, bytesSent uint64, bytesRecv uint64, t time.Time) error {
	record := IfaceRecord{
		Name:      name,
		BytesSent: bytesSent,
		BytesRecv: bytesRecv,
		Time:      t,
	}
	err := db.Create(&record).Error
	if err != nil {
		return err
	}

	return nil
}

func FindIfaceNewRecord(name string) (*IfaceRecord, error) {
	var res IfaceRecord

	err := db.Model(&IfaceRecord{}).Where("`name` = ?", name).Order("time desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &res, nil
}

func FindIfaceLastRecord(name string) (*IfaceRecord, error) {
	var res IfaceRecord
	err := db.Model(&IfaceRecord{}).Where("`name` = ?", name).Order("time asc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &res, nil
}

func FindIfaceRecord(name string, before time.Time) (*IfaceRecord, error) {
	var res IfaceRecord
	err := db.Model(&IfaceRecord{}).Where("`name` = ? AND `time` < ?", name, before).Order("time desc").First(&res).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return &res, nil
}

func CleanIfaceRecord(name string, keep time.Duration) error {
	dl := time.Now().Add(-1 * keep)
	err := db.Unscoped().Model(&IfaceRecord{}).Where("`name` = ? AND `time` < ?", name, dl).Delete(&IfaceRecord{}).Error
	if err != nil {
		return err
	}

	return nil
}
