package server

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/redisserver"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"github.com/pires/go-proxyproto"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	StatusContinue = "continue"
	StatusStop     = "stop"
)

type TcpServer struct {
	port       int64
	dest       string
	targetAddr *net.TCPAddr
	listener   *proxyproto.Listener
	swg        sync.WaitGroup
	allconn    sync.Map
	stopchan   chan bool
}

func NewTcpServer(port int64, dest string) (*TcpServer, error) {
	targetAddr, err := net.ResolveTCPAddr("tcp", dest)
	if err != nil {
		return nil, err
	}
	return &TcpServer{
		port:       port,
		dest:       dest,
		targetAddr: targetAddr,
		stopchan:   make(chan bool, 2),
	}, nil
}

func (t *TcpServer) Start() (err error) {
	if t.listener != nil {
		return nil
	}

	_listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.port))
	if err != nil {
		return fmt.Errorf("listen %d failed: %s", t.port, err.Error())
	}

	t.listener = &proxyproto.Listener{
		Listener: _listener,
	}

	logger.Infof("listen on %d start", t.port)

	go func() {
		defer func() {
			_ = t.listener.Close()
			t.listener = nil
		}()

	MainCycle:
		for {
			status := func() string {
				defer func() {
					if r := recover(); r != nil {
						if err, ok := r.(error); ok {
							logger.Panicf("listen on %d panic (error) : %s", t.port, err.Error())
						} else {
							logger.Panicf("listen on %d panic : %s", t.port, err.Error())
						}
					}
				}()

				select {
				case <-t.stopchan:
					return StatusStop
				default:
					// pass
				}

				conn, err := t.listener.Accept()
				if err != nil {
					logger.Errorf("listen on %d accecpt error: %s", t.port, err.Error())
					return StatusContinue
				}
				defer func() {
					if conn != nil {
						_ = conn.Close()
					}
				}()

				remoteAddr := conn.RemoteAddr()
				if remoteAddr == nil {
					return StatusContinue
				}

				if !t.remoteAddrCheck(remoteAddr.String()) {
					return StatusContinue
				}

				target, err := net.DialTCP("tcp", nil, t.targetAddr)
				if err != nil {
					logger.Errorf("Failed to connect to target %s: %v", t.dest, err)
					return StatusContinue
				}

				header := proxyproto.HeaderProxyFromAddrs(1, remoteAddr, t.targetAddr)
				_, err = header.WriteTo(target)
				if err != nil {
					logger.Errorf("Failed to write proxy header to target %s: %v", t.dest, err)
					return StatusContinue
				}

				_conn := conn
				conn = nil
				go t.forward(remoteAddr.String(), _conn, target)

				return StatusContinue
			}()
			if status == StatusStop {
				break MainCycle
			}
		}

		logger.Infof("listen on %d stop", t.port)
	}()

	return nil
}

func (t *TcpServer) Stop() error {
	if t.listener == nil {
		return nil
	}

	t.stopchan <- true
	time.Sleep(1 * time.Second)

	go func() {
		time.Sleep(time.Second * 10)
		t.allconn.Range(func(key, value any) bool {
			conn, ok := value.(net.Conn)
			if !ok {
				return true
			}

			_ = conn.Close()
			return true
		})
	}()

	t.swg.Wait()
	return nil
}

func (t *TcpServer) forward(remoteAddr string, conn net.Conn, target net.Conn) {
	t.swg.Add(1)
	defer t.swg.Done()

	if _, loaded := t.allconn.LoadOrStore(remoteAddr, conn); loaded {
		logger.Errorf("%s is already connected", remoteAddr)
		return
	}
	defer func() {
		t.allconn.Delete(remoteAddr)
	}()

	var stopchan = make(chan bool, 3)
	var wg sync.WaitGroup

	defer wg.Wait()

	defer func() {
		defer func() {
			_ = recover()
		}()

		_conn := conn
		conn = nil
		_ = _conn.Close()

	}()

	defer func() {
		defer func() {
			_ = recover()
		}()

		_target := target
		target = nil
		_ = _target.Close()
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					logger.Panicf("failed to forward from %s to %s: %s", conn.RemoteAddr(), target.RemoteAddr(), err.Error())
				} else {
					logger.Panicf("failed to forward from %s to %s: %v", conn.RemoteAddr(), target.RemoteAddr(), r)
				}
			}
		}()
		defer func() {
			stopchan <- true
		}()

		_, err := io.Copy(target, conn)
		if err != nil && conn != nil && target != nil && t.listener != nil {
			logger.Errorf("failed to forward from %s to %s: %v", conn.RemoteAddr(), target.RemoteAddr(), err)
		}
	}()

	go func() {
		wg.Add(2)
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					logger.Panicf("failed to forward from %s to %s: %s", target.RemoteAddr(), conn.RemoteAddr(), err.Error())
				} else {
					logger.Panicf("failed to forward from %s to %s: %v", target.RemoteAddr(), conn.RemoteAddr(), r)
				}
			}
		}()
		defer func() {
			stopchan <- true
		}()

		_, err := io.Copy(conn, target)
		if err != nil && conn != nil && target != nil && t.listener != nil {
			logger.Errorf("failed to forward from %s to %s: %v", target.RemoteAddr(), conn.RemoteAddr(), err)
		}
	}()

	<-stopchan
	return
}

func (t *TcpServer) remoteAddrCheck(remoteAddr string) bool {
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
