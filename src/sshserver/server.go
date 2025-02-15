package sshserver

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/ipcheck"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/pires/go-proxyproto"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type SshServer struct {
	status atomic.Int32
	config *config.SshForwardConfig

	ln4              net.Listener
	ln4Proxy         bool
	ln4Cross         bool
	ln4Target        *net.TCPAddr
	ln4TargetNetwork string

	ln6              net.Listener
	ln6Proxy         bool
	ln6Cross         bool
	ln6Target        *net.TCPAddr
	ln6TargetNetwork string

	swg        sync.WaitGroup
	allconn    sync.Map
	stopchan   chan bool
	controller SshController
}

type SshServerOpt struct {
	Config     *config.SshForwardConfig
	Controller SshController
}

func NewSshServer(opt *SshServerOpt) (*SshServer, error) {
	if opt.Config.ResolveIPv4DestAddress == nil && opt.Config.ResolveIPv6DestAddress == nil {
		return nil, fmt.Errorf("no dest address")
	}

	res := &SshServer{
		config:     opt.Config,
		controller: opt.Controller,
	}

	res.status.Store(StatusReady)

	return res, nil
}

func (s *SshServer) Start() (err error) {
	if s.ln4 != nil || s.status.Load() != StatusReady {
		return nil
	}

	if ipcheck.SupportIPv4() {
		if s.config.ResolveIPv4DestAddress != nil {
			_ln4, err := net.ListenTCP("tcp4", s.config.ResolveIPv4SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp4 failed: %s", s.config.SrcPort, err.Error())
			}

			if s.config.IPv4SrcServerProxy.IsEnable(true) {
				s.ln4 = &proxyproto.Listener{
					Listener: _ln4,
				}
				s.ln4Proxy = true
				s.ln4Cross = false
				s.ln4Target = s.config.ResolveIPv4DestAddress
				s.ln4TargetNetwork = "tcp4"
			} else {
				s.ln4 = _ln4
				s.ln4Proxy = false
				s.ln4Cross = false
				s.ln4Target = s.config.ResolveIPv4DestAddress
				s.ln4TargetNetwork = "tcp4"
			}
		} else if s.config.Cross && s.config.ResolveIPv6DestAddress != nil {
			_ln4, err := net.ListenTCP("tcp4", s.config.ResolveIPv4SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp4 failed: %s", s.config.SrcPort, err.Error())
			}

			s.ln4 = _ln4
			s.ln4Proxy = false
			s.ln4Cross = true
			s.ln4Target = s.config.ResolveIPv6DestAddress
			s.ln4TargetNetwork = "tcp6"
		}
	} else {
		s.ln4 = nil
		s.ln4Proxy = false
		s.ln4Cross = false
		s.ln4Target = nil
		s.ln4TargetNetwork = ""
	}

	if ipcheck.SupportIPv6() {
		if s.config.ResolveIPv6DestAddress != nil {
			_ln6, err := net.ListenTCP("tcp6", s.config.ResolveIPv6SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp6 failed: %s", s.config.SrcPort, err.Error())
			}

			if s.config.IPv6SrcServerProxy.IsEnable(true) {
				s.ln6 = &proxyproto.Listener{
					Listener: _ln6,
				}
				s.ln6Proxy = true
				s.ln6Cross = false
				s.ln6Target = s.config.ResolveIPv6DestAddress
				s.ln6TargetNetwork = "tcp6"
			} else {
				s.ln6 = _ln6
				s.ln6Proxy = false
				s.ln6Cross = false
				s.ln6Target = s.config.ResolveIPv6DestAddress
				s.ln6TargetNetwork = "tcp6"
			}
		} else if s.config.Cross && s.config.ResolveIPv4DestAddress != nil {
			_ln6, err := net.ListenTCP("tcp6", s.config.ResolveIPv6SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp6 failed: %s", s.config.SrcPort, err.Error())
			}

			s.ln6 = _ln6
			s.ln6Proxy = false
			s.ln6Cross = true
			s.ln6Target = s.config.ResolveIPv4DestAddress
			s.ln6TargetNetwork = "tcp4"
		}
	} else {
		s.ln6 = nil
		s.ln6Proxy = false
		s.ln6Cross = false
		s.ln6Target = nil
		s.ln6TargetNetwork = ""
	}

	if s.ln4 == nil && s.ln6 == nil {
		return fmt.Errorf("no listen address")
	}

	if s.ln4Target == nil && s.ln6Target == nil {
		return fmt.Errorf("no target address")
	}

	s.stopchan = make(chan bool, 4)

	if s.ln4 != nil {
		go func() {
			defer func() {
				_ = s.ln4.Close()
				s.ln4 = nil
			}()

			logger.Infof("listen on %d (ipv4) start", s.config.SrcPort)
		MainCycle:
			for {
				select {
				case <-s.stopchan:
					break MainCycle
				default:
					// pass
				}

				status := s.accept(s.ln4,
					"tcp4",
					!s.ln4Cross && s.config.IPv4DestRequestProxy.IsEnable(true),
					s.config.IPv4DestRequestProxyVersion,
					s.ln4TargetNetwork,
					s.ln4Target)
				if status == StatusStop {
					break MainCycle
				}
			}

			logger.Infof("listen on %d (ipv4) stop", s.config.SrcPort)
		}()
	}

	if s.ln6 != nil {
		go func() {
			defer func() {
				_ = s.ln6.Close()
				s.ln6 = nil
			}()

			logger.Infof("listen on %d (ipv6) start", s.config.SrcPort)
		MainCycle:
			for {
				select {
				case <-s.stopchan:
					break MainCycle
				default:
					// pass
				}

				status := s.accept(s.ln6,
					"tcp6",
					!s.ln6Cross && s.config.IPv6DestRequestProxy.IsEnable(true),
					s.config.IPv6DestRequestProxyVersion,
					s.ln6TargetNetwork,
					s.ln6Target)
				if status == StatusStop {
					break MainCycle
				}
			}

			logger.Infof("listen on %d (ipv6) stop", s.config.SrcPort)
		}()
	}

	if !s.status.CompareAndSwap(StatusReady, StatusRunning) {
		return fmt.Errorf("server run failed: can not set status")
	}

	return nil
}

func (s *SshServer) Stop() error {
	if s.ln4 == nil || !s.status.CompareAndSwap(StatusRunning, StatusStopping) {
		return nil
	}

	s.stopchan <- true
	s.stopchan <- true
	close(s.stopchan)
	time.Sleep(1 * time.Second)

	go func() {
		time.Sleep(time.Second * 10)
		s.allconn.Range(func(key, value any) bool {
			conn, ok := value.(net.Conn)
			if !ok {
				return true
			}

			_ = conn.Close()
			return true
		})
		s.allconn.Clear()
	}()

	s.swg.Wait()

	s.status.CompareAndSwap(StatusStopping, StatusFinished)
	return nil
}

func (s *SshServer) forward(remoteAddr string, conn net.Conn, target net.Conn, record *database.SshConnectRecord) {
	defer func() {
		defer func() {
			_ = recover()
		}()

		err := database.UpdateSshConnectRecord(record, "连接正常断开。")
		if err != nil {
			logger.Errorf("update ssh connect record error: %s", record)
		}
	}()

	defer func() {
		r := recover()
		if r != nil {
			if err, ok := r.(error); ok {
				logger.Panicf("ssh forward panic error: %s", err.Error())
			} else {
				logger.Panicf("ssh forward panic error: %v", r)
			}
		}
	}()

	s.swg.Add(1)
	defer s.swg.Done()

	if _, loaded := s.allconn.LoadOrStore(remoteAddr, conn); loaded {
		logger.Errorf("%s is already connected", remoteAddr)
		return
	}
	defer func() {
		s.allconn.Delete(remoteAddr)
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
					logger.Panicf("failed to forward: %s", err.Error())
				} else {
					logger.Panicf("failed to forward: %v", r)
				}
			}
		}()

		defer func() {
			stopchan <- true
		}()

		_, err := io.Copy(target, conn)
		if err != nil && conn != nil && target != nil && s.status.Load() == StatusRunning {
			logger.Errorf("failed to forward from %s to %s: %v", conn.RemoteAddr(), target.RemoteAddr(), err)
		}
	}()

	go func() {
		wg.Add(1)
		defer wg.Done()

		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					logger.Panicf("failed to forward from: %s", err.Error())
				} else {
					logger.Panicf("failed to forward from: %v", r)
				}
			}
		}()

		defer func() {
			stopchan <- true
		}()

		_, err := io.Copy(conn, target)
		if err != nil && conn != nil && target != nil && s.status.Load() == StatusRunning {
			logger.Errorf("failed to forward from %s to %s: %v", target.RemoteAddr(), conn.RemoteAddr(), err)
		}
	}()

	<-stopchan
	return
}

func (s *SshServer) accept(ln net.Listener, srcNetwork string, destProxy bool, destProxyVersion int, targetNetwork string, targetAddr *net.TCPAddr) string {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				logger.Panicf("listen on %d panic (error) : %s", s.config.SrcPort, err.Error())
			} else {
				logger.Panicf("listen on %d panic : %s", s.config.SrcPort, err.Error())
			}
		}
	}()

	if ln == nil {
		return StatusStop
	}

	conn, err := ln.Accept()
	if err != nil {
		logger.Errorf("listen on %d accecpt error: %s", s.config.SrcPort, err.Error())
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

	remoteSSHAddr, err := net.ResolveTCPAddr(srcNetwork, remoteAddr.String())
	if err != nil {
		return StatusContinue
	}

	ckErr := s.controller.RemoteAddrCheck(remoteSSHAddr, targetAddr, s.config.CountRules)
	if ckErr != nil {
		_, _ = database.AddSshConnectRecord("", remoteSSHAddr.IP, targetAddr, false, time.Now(), fmt.Sprintf("来访IP检查出现问题。%s", ckErr.Error()))
		return StatusContinue
	}

	target, err := net.DialTCP(targetNetwork, nil, targetAddr)
	if err != nil {
		logger.Errorf("Failed to connect to target %s: %v", targetAddr.String(), err)
		_, _ = database.AddSshConnectRecord("", remoteSSHAddr.IP, targetAddr, false, time.Now(), "无法解析来访TCP地址。")
		return StatusContinue
	}
	defer func() {
		if target != nil {
			_ = target.Close()
		}
	}()

	if destProxy {
		header := proxyproto.HeaderProxyFromAddrs(byte(destProxyVersion), remoteSSHAddr, targetAddr)
		_, err = header.WriteTo(target)
		if err != nil {
			logger.Errorf("Failed to write proxy header to target %s: %v", targetAddr.String(), err)
			_, _ = database.AddSshConnectRecord("", remoteSSHAddr.IP, targetAddr, false, time.Now(), "无法写入Proxy协议头部。")
			return StatusContinue
		}
	}

	record, err := database.AddSshConnectRecord("", remoteSSHAddr.IP, targetAddr, true, time.Now(), "允许建立连接。")
	if err != nil {
		logger.Errorf("Fail to save ssh connect record to database: %s", err.Error())
		_, _ = database.AddSshConnectRecord("", remoteSSHAddr.IP, targetAddr, true, time.Now(), "无法记录SSH数据，不允许建立连接。")
		return StatusContinue
	}

	_conn := conn
	_target := target
	conn = nil
	target = nil
	go s.forward(remoteAddr.String(), _conn, _target, record)

	return StatusContinue
}
