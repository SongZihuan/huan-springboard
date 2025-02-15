package sshserver

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/api/apiip"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/redisserver"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var sshServerGroupOnce sync.Once
var sshServerGroup *SshServerGroup

type SshServerGroup struct {
	status  atomic.Int32
	servers sync.Map
}

func NewSshServerGroup() (res *SshServerGroup) { // 单例模式
	sshServerGroupOnce.Do(func() {
		sshServerGroup = &SshServerGroup{}
		sshServerGroup.status.Store(StatusReady)
	})
	return sshServerGroup
}

func (s *SshServerGroup) Start() error {
	if !s.status.CompareAndSwap(StatusReady, StatusWaitStart) {
		return nil
	}

	err := s.StartAllServers()
	if err != nil {
		return err
	}

	return nil
}

func (s *SshServerGroup) StartAllServers() error {
	if !s.status.CompareAndSwap(StatusWaitStart, StatusRunning) {
		return nil
	}

	logger.Infof("SSH ServerGroup All Server Start...")
	for _, f := range config.GetConfig().SSH.Forward {
		server, err := NewSshServer(&SshServerOpt{
			Config:     f,
			Controller: s,
		})
		if err != nil {
			logger.Errorf("New SSH Server Error: %s\n", err)
		}

		_, loaded := s.servers.LoadOrStore(f.SrcPort, server)
		if loaded {
			logger.Errorf("SSH Port Conflict: %d\n", f.SrcPort)
		}

		err = server.Start()
		if err != nil {
			logger.Errorf("Start SSH Server Error: %s\n", err)
			// 没启动成功，但仍然保留在 Map 中，目的是提前发现可能的端口冲突（配置错误）
		}
	}
	logger.Infof("SSH ServerGroup All Server Start Finished")

	return nil
}

func (s *SshServerGroup) Stop() error {
	_ = s.StopAllServers()

	if !s.status.CompareAndSwap(StatusWaitStop, StatusStopping) {
		return nil
	}

	s.status.CompareAndSwap(StatusStopping, StatusFinished)
	return nil
}

func (s *SshServerGroup) StopAllServers() error {
	if !s.status.CompareAndSwap(StatusRunning, StatusWaitStop) {
		return nil
	}

	var wg sync.WaitGroup

	logger.Infof("SSH ServerGroup All Server Stop...")
	s.servers.Range(func(key, value any) bool {
		server, ok := value.(*SshServer)
		if !ok {
			return true
		}

		go func(server *SshServer) {
			wg.Add(1)
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						logger.Panicf("stop ssh server panic error: %s\n", err.Error())
					} else {
						logger.Panicf("stop ssh server panic: %v\n", r)
					}
				}
			}()

			_ = server.Stop()
		}(server)

		return true
	})
	s.servers.Clear()

	wg.Wait()
	logger.Infof("SSH ServerGroup All Server Stop Finished")

	s.status.CompareAndSwap(StatusStopping, StatusFinished)

	return nil
}

func (s *SshServerGroup) RestartAllServers() error {
	if !s.status.CompareAndSwap(StatusWaitStop, StatusWaitStart) {
		return fmt.Errorf("can not restart ssh server")
	}

	err := s.StartAllServers()
	if err != nil {
		return err
	}

	return nil
}

func (s *SshServerGroup) RemoteAddrCheck(remoteAddr *net.TCPAddr, to *net.TCPAddr, countRules []*config.SshCountRuleConfig) error {
	ip := remoteAddr.IP
	if ip == nil {
		return fmt.Errorf("无法获取IP")
	}

	if len(countRules) == 0 {
		countRules = config.GetConfig().SSH.RuleList.CountRules
	}

	if ip.IsLoopback() && (config.GetConfig().SSH.RuleList.AlwaysAllowIntranet.IsEnable(false) || config.GetConfig().SSH.RuleList.AlwaysAllowLoopback.IsEnable(true)) {
		return nil
	}

	if !database.SshCheckIP(ip.String()) {
		return fmt.Errorf("IP被SQLite封禁。")
	}

	if ip.IsPrivate() && config.GetConfig().SSH.RuleList.AlwaysAllowIntranet.IsEnable(false) {
		return nil
	}

	rcErr := s.CountRulesCheck(ip, to, countRules)
	if rcErr != nil {
		return rcErr
	}

	var loc *apiip.QueryIpLocationData = nil
	if !ip.IsPrivate() && !ip.IsLoopback() {
		var err error

		loc, err = redisserver.QueryIpLocation(ip.String())
		if err != nil || loc == nil || strings.Contains(loc.Isp, "专用网络") || strings.Contains(loc.Isp, "本地环回") || strings.Contains(loc.Isp, "本地回环") {
			if err != nil {
				logger.Errorf("failed to query ip location: %s", err.Error())
			} else if loc == nil {
				logger.Panicf("failed to query ip location: loc is nil")
			}

			loc = nil
		} else {
			if !database.TcpCheckLocationNation(loc.Nation) ||
				!database.TcpCheckLocationProvince(loc.Province) ||
				!database.TcpCheckLocationCity(loc.City) ||
				!database.TcpCheckLocationISP(loc.Isp) {
				return fmt.Errorf("IP地址被SQLite封禁。")
			}
		}
	} else {
		loc = nil
	}

RuleCycle:
	for _, r := range config.GetConfig().SSH.RuleList.RuleList {
		if loc == nil {
			if r.HasLocation() {
				continue RuleCycle
			}
		} else {
			ok, err := loc.CheckLocation(&r.RuleConfig)
			if err != nil {
				logger.Errorf("check location error: %s", err.Error())
				return fmt.Errorf("在配置文件规则策略中，检测IP地址错误。")
			} else if !ok {
				continue RuleCycle
			}
		}

		ok, err := r.CheckIP(ip)
		if err != nil {
			logger.Errorf("check ip error: %s", err.Error())
			return fmt.Errorf("在配置文件规则策略中，检测IP信息错误。")
		} else if !ok {
			continue RuleCycle
		}

		if r.Banned.ToBool(true) { // true - 封禁
			return fmt.Errorf("IP在配置文件规则策略中被封禁。")
		}

		return nil
	}

	if config.GetConfig().SSH.RuleList.DefaultBanned.ToBool(true) { // true - 封禁
		return fmt.Errorf("IP在配置文件默认兜底规则策略中被封禁。")
	}

	return nil
}

func (s *SshServerGroup) CountRulesCheck(ip net.IP, to *net.TCPAddr, countRules []*config.SshCountRuleConfig) error {
	now := time.Now()

	if !redisserver.QuerySSHIpBanned(ip.String()) {
		return fmt.Errorf("IP在配置文件计数策略中被封禁，IP已被Redis封禁。")
	}

	if len(countRules) > 0 {
		limit := int(countRules[0].TryCount + 1) // +1防止TryCount是0
		after := now.Add(-1 * time.Second * time.Duration(countRules[0].MemorySeconds))

		res, err := database.FindSshConnectRecord("", ip, to, limit, after)
		if err != nil {
			logger.Errorf("count rules check error: %s", err.Error())
			return fmt.Errorf("从数据库读取SSH记录异常，禁止连接。")
		}

		for _, r := range countRules {
			if s._countRulesCheck(res, r, now) {
				if r.BannedSeconds <= 0 {
					return nil // 返回是否放行，true表示放行
				}

				err := redisserver.SetSSHIpBanned(ip.String(), time.Duration(r.BannedSeconds)*time.Second)
				if err != nil {
					logger.Errorf("count rules check error: %s", err.Error())
				}
				return fmt.Errorf("IP在配置文件计数策略中被封禁, 时长 %d 秒。", r.BannedSeconds)
			}
		}
	} else {
		// 默认策略：3分钟内5次以上, 封禁10分钟
		limit := 10                              // +1防止TryCount是0
		after := now.Add(-1 * time.Second * 180) // 三分钟

		res, err := database.FindSshConnectRecord("", ip, to, limit, after)
		if err != nil {
			logger.Errorf("count rules check error: %s", err.Error())
			return fmt.Errorf("从数据库读取SSH记录异常，禁止连接。")
		}

		if len(res) > 5 {
			// 命中默认策略
			err := redisserver.SetSSHIpBanned(ip.String(), 600*time.Second)
			if err != nil {
				logger.Errorf("count rules check error: %s", err.Error())
			}
			return fmt.Errorf("IP在配置文件计数策略中被封禁, 时长 %d 秒。", 600)
		}
	}

	return nil // 没有命中封禁策略
}

func (*SshServerGroup) _countRulesCheck(record []database.SshConnectRecord, rules *config.SshCountRuleConfig, now time.Time) bool {
	var index = 0
	after := now.Add(-1 * time.Second * time.Duration(rules.MemorySeconds))

	for i, r := range record {
		if r.Time.After(after) {
			index = i
			break
		}
	}

	return len(record)-index > int(rules.TryCount) // 返回是否命中策略，true表示命中 (使用大于, 而不是大于等于)
}
