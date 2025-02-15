package config

import (
	"github.com/SongZihuan/huan-springboard/src/utils"
	"time"
)

type DBCleanConfig struct {
	ExecutionIntervalHour int64 `yaml:"execution-interval-hour"`

	IfaceRecordSaveRetentionPeriod string `yaml:"iface-record-save-retention-period"`
	SSHRecordSaveRetentionPeriod   string `yaml:"ssh-record-save-retention-period"`

	IfaceRecordSaveTime time.Duration `yaml:"-"`
	SSHRecordSaveTime   time.Duration `yaml:"-"`
}

func (d *DBCleanConfig) setDefault() {
	if d.ExecutionIntervalHour <= 0 {
		d.ExecutionIntervalHour = 6
	}

	if d.IfaceRecordSaveRetentionPeriod == "" {
		d.IfaceRecordSaveRetentionPeriod = "3M"
	}

	if d.SSHRecordSaveRetentionPeriod == "" {
		d.SSHRecordSaveRetentionPeriod = "3M"
	}

	return
}

func (d *DBCleanConfig) check() (err ConfigError) {
	d.IfaceRecordSaveTime = utils.ReadTimeDuration(d.IfaceRecordSaveRetentionPeriod)
	d.SSHRecordSaveTime = utils.ReadTimeDuration(d.SSHRecordSaveRetentionPeriod)

	if d.IfaceRecordSaveTime == 0 {
		return NewConfigError("bad iface-record-save-retention-period")
	}

	if d.SSHRecordSaveTime == 0 {
		return NewConfigError("bad ssh-record-save-retention-period")
	}

	if d.IfaceRecordSaveTime == -1 {
		_ = NewConfigWarning("iface-record-save-retention-period is set to be saved permanently")
	} else if d.IfaceRecordSaveTime < time.Minute*5 {
		return NewConfigError("bad iface-record-save-retention-period, must more than 5 minute")
	}

	if d.SSHRecordSaveTime == -1 {
		_ = NewConfigWarning("ssh-record-save-retention-period is set to be saved permanently")
	} else if d.SSHRecordSaveTime < time.Minute*5 {
		return NewConfigError("bad ssh-record-save-retention-period, must more than 5 minute")
	}

	return nil
}
