package config

type SshCountRuleConfig struct {
	TryCount      int64 `yaml:"try-count"`      // 尝试次数
	MemorySeconds int64 `yaml:"memory-seconds"` // 记录保持时间
	BannedSeconds int64 `yaml:"banned-seconds"` // 封禁时长
}

func (s *SshCountRuleConfig) setDefault() {
	return
}

func (s *SshCountRuleConfig) check() (err ConfigError) {
	if s.TryCount < 0 {
		s.TryCount = 0
	}

	if s.MemorySeconds <= 0 {
		return NewConfigError("memory-seconds must be greater than 0")
	}

	if s.BannedSeconds <= 0 {
		return NewConfigError("banned-seconds must be greater than 0")
	}
	return nil
}
