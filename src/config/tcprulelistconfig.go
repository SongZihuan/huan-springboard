package config

import (
	"github.com/SongZihuan/huan-springboard/src/network"
	"github.com/SongZihuan/huan-springboard/src/utils"
)

type TcpRuleListConfig struct {
	RuleList            []*TcpRuleConfig `yaml:"rules"`
	DefaultBanned       utils.StringBool `yaml:"default-banned"`        // 默认（未名字规则）拒绝连接
	AlwaysAllowIntranet utils.StringBool `yaml:"always-allow-intranet"` // 总是允许内网连接（配置 ip 数据库封禁除外）
	AlwaysAllowLoopback utils.StringBool `yaml:"always-allow-loopback"` // 总是允许本地回环地址连接（不检查 ip 数据库封禁）

	InterfaceName              string `yaml:"interface-name"`                 // 监听的网卡名
	DataCollectionCycleSeconds uint64 `yaml:"data-collection-cycle-seconds"`  // 数据收集周期：建议5s
	StatisticalTimeSpanSeconds uint64 `yaml:"statistical-time-span-seconds"`  // 数据统计时间跨度，单位秒（计算平均值时，时间的跨度。例如获取5分钟内接受到的数据包，然后除以5，得到每秒平均bytes，供下文使用
	StatisticalPeriodSeconds   uint64 `yaml:"statistical-period-seconds"`     // 数据统计周期，多长时间进行一次数据统计，以及给出是否启用限流
	ReceiveBytesOfCycle        string `yaml:"receive-bytes-of-cycle"`         // 入网流量限制（单位Bytes/S), 0 表示不i按照
	TransmitBytesOfCycle       string `yaml:"transmit-bytes-of-cycle"`        // 出网流量限制（单位Bytes/S), 0 表示不限制
	StopAcceptTimeLimitSeconds uint64 `yaml:"stop-accept-time-limit-seconds"` // 高负荷多久后关停服务

	SentLimit uint64 `yaml:"-"`
	RecvLimit uint64 `yaml:"-"`
}

func (t *TcpRuleListConfig) setDefault() {
	for _, r := range t.RuleList {
		r.setDefault()
	}

	t.DefaultBanned.SetDefaultDisable()
	t.AlwaysAllowIntranet.SetDefaultEnable()
	t.AlwaysAllowLoopback.SetDefaultEnable()

	if t.InterfaceName != "" {
		if t.DataCollectionCycleSeconds <= 0 {
			t.DataCollectionCycleSeconds = 5
		}

		if t.StatisticalTimeSpanSeconds <= 0 {
			t.StatisticalTimeSpanSeconds = 1800 // 30分钟
		}

		if t.StatisticalPeriodSeconds <= 0 {
			t.StatisticalPeriodSeconds = 10
		}

		if t.StopAcceptTimeLimitSeconds <= 0 {
			t.StopAcceptTimeLimitSeconds = 3600 // 1小时
		}

		if t.ReceiveBytesOfCycle == "" {
			t.ReceiveBytesOfCycle = "0"
		}

		if t.TransmitBytesOfCycle == "" {
			t.TransmitBytesOfCycle = "0"
		}
	}

	return
}

func (t *TcpRuleListConfig) check() (err ConfigError) {
	for _, r := range t.RuleList {
		err := r.check()
		if err != nil && err.IsError() {
			return err
		}
	}

	if t.InterfaceName != "" {
		if _, ok := network.Iface[t.InterfaceName]; !ok {
			return NewConfigError("bad interface name")
		}

		t.SentLimit = utils.ReadBytes(t.TransmitBytesOfCycle)
		t.RecvLimit = utils.ReadBytes(t.ReceiveBytesOfCycle)
	} else {
		t.SentLimit = 0
		t.RecvLimit = 0
	}

	return nil
}
