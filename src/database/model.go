package database

import (
	"database/sql"
)

// Model gorm.Model的仿写，明确了键名
type Model struct {
	ID uint `gorm:"column:id;primarykey"`
}

type BannedIP struct {
	Model
	IP      string       `gorm:"column:ip;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*BannedIP) TableName() string {
	return "banned_ip"
}

type BannedLocationNation struct {
	Model
	Nation  sql.NullString `gorm:"column:nation;"`
	StartAt sql.NullTime   `gorm:"column:start_at;"`
	StopAt  sql.NullTime   `gorm:"column:stop_at;"`
}

func (*BannedLocationNation) TableName() string {
	return "banned_location_nation"
}

type BannedLocationProvince struct {
	Model
	Province sql.NullString `gorm:"column:province;"`
	StartAt  sql.NullTime   `gorm:"column:start_at;"`
	StopAt   sql.NullTime   `gorm:"column:stop_at;"`
}

func (*BannedLocationProvince) TableName() string {
	return "banned_location_province"
}

type BannedLocationCity struct {
	Model
	City    sql.NullString `gorm:"column:city;"`
	StartAt sql.NullTime   `gorm:"column:start_at;"`
	StopAt  sql.NullTime   `gorm:"column:stop_at;"`
}

func (*BannedLocationCity) TableName() string {
	return "banned_location_city"
}

type BannedLocationISP struct {
	Model
	ISP     sql.NullString `gorm:"column:isp;"`
	StartAt sql.NullTime   `gorm:"column:start_at;"`
	StopAt  sql.NullTime   `gorm:"column:stop_at;"`
}

func (*BannedLocationISP) TableName() string {
	return "banned_location_isp"
}
