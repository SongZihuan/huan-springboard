package database

import (
	"database/sql"
	"time"
)

// Model gorm.Model的仿写，明确了键名
type Model struct {
	ID uint `gorm:"column:id;primarykey"`
}

type TcpBannedIP struct {
	Model
	IP      string       `gorm:"column:ip;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*TcpBannedIP) TableName() string {
	return "tcp_banned_ip"
}

type TcpBannedLocationNation struct {
	Model
	Nation  string       `gorm:"column:nation;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*TcpBannedLocationNation) TableName() string {
	return "tcp_banned_location_nation"
}

type TcpBannedLocationProvince struct {
	Model
	Province string       `gorm:"column:province;type:VARCHAR(50);not null;"`
	StartAt  sql.NullTime `gorm:"column:start_at;"`
	StopAt   sql.NullTime `gorm:"column:stop_at;"`
}

func (*TcpBannedLocationProvince) TableName() string {
	return "tcp_banned_location_province"
}

type TcpBannedLocationCity struct {
	Model
	City    string       `gorm:"column:city;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*TcpBannedLocationCity) TableName() string {
	return "tcp_banned_location_city"
}

type TcpBannedLocationISP struct {
	Model
	ISP     string       `gorm:"column:isp;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*TcpBannedLocationISP) TableName() string {
	return "tcp_banned_location_isp"
}

type SshBannedIP struct {
	Model
	IP      string       `gorm:"column:ip;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*SshBannedIP) TableName() string {
	return "ssh_banned_ip"
}

type SshBannedLocationNation struct {
	Model
	Nation  string       `gorm:"column:nation;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*SshBannedLocationNation) TableName() string {
	return "ssh_banned_location_nation"
}

type SshBannedLocationProvince struct {
	Model
	Province string       `gorm:"column:province;type:VARCHAR(50);not null;"`
	StartAt  sql.NullTime `gorm:"column:start_at;"`
	StopAt   sql.NullTime `gorm:"column:stop_at;"`
}

func (*SshBannedLocationProvince) TableName() string {
	return "ssh_banned_location_province"
}

type SshBannedLocationCity struct {
	Model
	City    string       `gorm:"column:city;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*SshBannedLocationCity) TableName() string {
	return "ssh_banned_location_city"
}

type SshBannedLocationISP struct {
	Model
	ISP     string       `gorm:"column:isp;type:VARCHAR(50);not null;"`
	StartAt sql.NullTime `gorm:"column:start_at;"`
	StopAt  sql.NullTime `gorm:"column:stop_at;"`
}

func (*SshBannedLocationISP) TableName() string {
	return "ssh_banned_location_isp"
}

type SshConnectRecord struct {
	Model
	From          string        `gorm:"column:from;type:VARCHAR(50);not null;"`
	To            string        `gorm:"column:to;type:VARCHAR(50);not null;"`
	Accept        bool          `gorm:"column:accept;not null;"`
	Time          time.Time     `gorm:"column:time;not null;"`
	TimeConsuming sql.NullInt64 `gorm:"column:time_consuming;"` // 单位：毫秒（Millisecond）
	Mark          string        `gorm:"column:mark;type:VARCHAR(200);not null;"`
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
