package database

import (
	"database/sql"
	"time"
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
	Nation  string       `gorm:"column:nation;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*BannedLocationNation) TableName() string {
	return "banned_location_nation"
}

type BannedLocationProvince struct {
	Model
	Province string       `gorm:"column:province;type:VARCHAR(50);not null;"`
	StartAt  sql.NullTime `gorm:"column:start_at;"`
	StopAt   sql.NullTime `gorm:"column:stop_at;"`
}

func (*BannedLocationProvince) TableName() string {
	return "banned_location_province"
}

type BannedLocationCity struct {
	Model
	City    string       `gorm:"column:city;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*BannedLocationCity) TableName() string {
	return "banned_location_city"
}

type BannedLocationISP struct {
	Model
	ISP     string       `gorm:"column:isp;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*BannedLocationISP) TableName() string {
	return "banned_location_isp"
}

type IfaceRecord struct {
	Model
	Name      string    `gorm:"column:name;VARCHAR(50);not null;"`
	BytesSent uint64    `gorm:"column:bytes_sent;not null;"`
	BytesRecv uint64    `gorm:"column:bytes_received;not null;"`
	Time      time.Time `gorm:"column:time;not null;"`
}

func (*IfaceRecord) TableName() string {
	return "iface_record"
}
