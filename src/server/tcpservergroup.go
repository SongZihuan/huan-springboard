package server

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/iface"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/redisserver"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"math"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var tcpServerGroupOnce sync.Once
var tcpServerGroup *TcpServerGroup

type TcpServerGroup struct {
	status              atomic.Int32
	watcher             *iface.NetWatcher
	ifaceNotify         chan *iface.NotifyData
	ifaceNotifyStopchan chan bool
	servers             sync.Map
	stopAccept          atomic.Bool
	stopAcceptTime      *time.Time // 仅限一个协程使用，因此不需要
}

func NewTcpServerGroup(watcher *iface.NetWatcher) (res *TcpServerGroup) { // 单例模式
	tcpServerGroupOnce.Do(func() {
		tcpServerGroup = &TcpServerGroup{
			watcher:     watcher,
			ifaceNotify: watcher.AddNotice("TcpServerGroup"),
		}
		tcpServerGroup.status.Store(StatusReady)
		tcpServerGroup.stopAccept.Store(false)
	})
	return tcpServerGroup
}

func (t *TcpServerGroup) Start() error {
	if !t.status.CompareAndSwap(StatusReady, StatusWaitStart) {
		return nil
	}

	t.ifaceNotifyStopchan = make(chan bool, 2)
	t.processIfaceNotify()

	err := t.StartAllServers()
	if err != nil {
		return err
	}

	return nil
}

func (t *TcpServerGroup) StartAllServers() error {
	if !t.status.CompareAndSwap(StatusWaitStart, StatusRunning) {
		return nil
	}

	logger.Infof("TCP ServerGroup All Server Start...")
	for _, f := range config.GetConfig().TCP.Forward {
		server, err := NewTcpServer(&TcpServerOpt{
			port:             f.SrcPort,
			dest:             f.DestAddress,
			srcProxy:         f.SrcServerProxy.IsEnable(true),
			destProxy:        f.DestRequestProxy.IsEnable(true),
			destProxyVersion: f.DestRequestProxyVersion,
			controller:       t,
		})
		if err != nil {
			logger.Errorf("New TCP Server Error: %s\n", err)
		}

		_, loaded := t.servers.LoadOrStore(f.SrcPort, server)
		if loaded {
			logger.Errorf("TCP Port Conflict: %d\n", f.SrcPort)
		}

		err = server.Start()
		if err != nil {
			logger.Errorf("Start TCP Server Error: %s\n", err)
			// 没启动成功，但仍然保留在 Map 中，目的是提前发现可能的端口冲突（配置错误）
		}
	}
	logger.Infof("TCP ServerGroup All Server Start Finished")

	return nil
}

func (t *TcpServerGroup) Stop() error {
	_ = t.StopAllServers()

	if !t.status.CompareAndSwap(StatusWaitStop, StatusStopping) {
		return nil
	}

	t.ifaceNotifyStopchan <- true
	close(t.ifaceNotifyStopchan)

	t.status.CompareAndSwap(StatusStopping, StatusFinished)
	return nil
}

func (t *TcpServerGroup) StopAllServers() error {
	if !t.status.CompareAndSwap(StatusRunning, StatusWaitStop) {
		return nil
	}

	var wg sync.WaitGroup

	logger.Infof("TCP ServerGroup All Server Stop...")
	t.servers.Range(func(key, value any) bool {
		server, ok := value.(*TcpServer)
		if !ok {
			return true
		}

		go func(server *TcpServer) {
			wg.Add(1)
			defer wg.Done()

			defer func() {
				if r := recover(); r != nil {
					if err, ok := r.(error); ok {
						logger.Panicf("stop server panic error: %s\n", err.Error())
					} else {
						logger.Panicf("stop server panic: %v\n", r)
					}
				}
			}()

			_ = server.Stop()
		}(server)

		return true
	})
	t.servers.Clear()

	wg.Wait()
	logger.Infof("TCP ServerGroup All Server Stop Finished")

	t.status.CompareAndSwap(StatusStopping, StatusFinished)

	return nil
}

func (t *TcpServerGroup) RestartAllServers() error {
	if !t.status.CompareAndSwap(StatusWaitStop, StatusWaitStart) {
		return fmt.Errorf("can not restart server")
	}

	err := t.StartAllServers()
	if err != nil {
		return err
	}

	return nil
}

func (t *TcpServerGroup) processIfaceNotify() {
	go func() {
	MainCycle:
		for {
			select {
			case data := <-t.ifaceNotify:
				if data.IsStop {
					break MainCycle
				}

				if uint64(math.Ceil(data.SpanOfSecond)) < min(30, config.GetConfig().TCP.RuleList.StatisticalTimeSpanSeconds) {
					continue MainCycle
				}

				fmt.Println("TAG A", config.GetConfig().TCP.RuleList.StopAcceptTimeLimitSeconds)

				// 要做为关闭 accept 的依据只需要满足 span 大于 min(30, config.GetConfig().TCP.RuleList.StatisticalTimeSpanSeconds)
				if (data.RecvLimit > 0 && !data.IsRecvOK) || (data.SentLimit > 0 && !data.IsSentOK) {
					t.stopAccept.Store(true)

					if t.stopAcceptTime == nil {
						ti := time.Now()
						t.stopAcceptTime = &ti
					} else if t.stopAcceptTime.Add(time.Duration(config.GetConfig().TCP.RuleList.StopAcceptTimeLimitSeconds) * time.Second).Before(time.Now()) {
						// 启用清理
						if t.status.Load() == StatusRunning {
							go func() {
								_ = t.StopAllServers()
							}()
							time.Sleep(1 * time.Second)
						}
					}

					continue MainCycle
				}

				// 剩余事件（else）就只有：data.IsRecvOK && data.IsSentOK
				// 要想开启 accept 则必须要稳定要 UseRealLastRecord, 也就是 span 要达到指定长度
				if data.UseRealLastRecord {
					if t.status.Load() == StatusWaitStop {
						go func() {
							err := t.RestartAllServers()
							if err != nil {
								logger.Errorf("TCP Server Group Restart All Server Failed: %s", err.Error())
							}
						}()
						time.Sleep(1 * time.Second)
					}

					t.stopAcceptTime = nil
					t.stopAccept.Store(false)
					continue MainCycle
				}

				// 其他数据排除
			case <-t.ifaceNotifyStopchan:
				break MainCycle
			}
		}

		logger.Infof("TCP ServerGroup Interface process stop")
	}()
}

func (t *TcpServerGroup) TcpNetworkAccept() bool {
	return !t.stopAccept.Load()
}

func (*TcpServerGroup) RemoteAddrCheck(remoteAddr string) bool {
	ipStr, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	if !database.CheckIP(ip.String()) {
		return false
	}

	if utils.IsPrivateIP(ip.String(), true) && config.GetConfig().TCP.RuleList.AlwaysAllIntranet.IsEnable(true) {
		return true
	}

	loc, err := redisserver.QueryIpLocation(ip.String())
	if err != nil || loc == nil || strings.Contains(loc.Isp, "专用网络") || strings.Contains(loc.Isp, "本地环回") || strings.Contains(loc.Isp, "本地回环") {
		if err != nil {
			logger.Errorf("failed to query ip location: %s", err.Error())
		} else if loc == nil {
			logger.Panicf("failed to query ip location: loc is nil")
		}

		loc = nil
	} else {
		if !database.CheckLocationNation(loc.Nation) ||
			!database.CheckLocationProvince(loc.Province) ||
			!database.CheckLocationCity(loc.City) ||
			!database.CheckLocationISP(loc.Isp) {
			return false
		}
	}

RuleCycle:
	for _, r := range config.GetConfig().TCP.RuleList.RuleList {
		if loc == nil {
			if r.HasLocation() {
				continue RuleCycle
			}
		} else {
			ok, err := loc.CheckLocation(&r.RuleConfig)
			if err != nil {
				logger.Errorf("check location error: %s", err.Error())
				return false
			} else if !ok {
				continue RuleCycle
			}
		}

		ok, err := r.CheckIP(ip)
		if err != nil {
			logger.Errorf("check ip error: %s", err.Error())
			return false
		} else if !ok {
			continue RuleCycle
		}

		return !r.Banned.ToBool(true) // Banned表示封禁，该函数（IPCheck）返回值表示允许通行，因此取反
	}

	return !config.GetConfig().TCP.RuleList.DefaultBanned.ToBool(false)
}
