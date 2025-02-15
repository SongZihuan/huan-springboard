package tcpserver

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/ipcheck"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/pires/go-proxyproto"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TcpServer struct {
	status atomic.Int32
	config *config.TcpForwardConfig

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
	controller TcpController
}

type TcpServerOpt struct {
	Config     *config.TcpForwardConfig
	Controller TcpController
}

func NewTcpServer(opt *TcpServerOpt) (*TcpServer, error) {
	if opt.Config.ResolveIPv4DestAddress == nil && opt.Config.ResolveIPv6DestAddress == nil {
		return nil, fmt.Errorf("no dest address")
	}

	res := &TcpServer{
		config:     opt.Config,
		controller: opt.Controller,
	}

	res.status.Store(StatusReady)

	return res, nil
}

func (t *TcpServer) Start() (err error) {
	if t.ln4 != nil || t.status.Load() != StatusReady {
		return nil
	}

	if ipcheck.SupportIPv4() {
		if t.config.ResolveIPv4DestAddress != nil {
			_ln4, err := net.ListenTCP("tcp4", t.config.ResolveIPv4SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp4 failed: %s", t.config.SrcPort, err.Error())
			}

			if t.config.IPv4SrcServerProxy.IsEnable(true) {
				t.ln4 = &proxyproto.Listener{
					Listener: _ln4,
				}
				t.ln4Proxy = true
				t.ln4Cross = false
				t.ln4Target = t.config.ResolveIPv4DestAddress
				t.ln4TargetNetwork = "tcp4"
			} else {
				t.ln4 = _ln4
				t.ln4Proxy = false
				t.ln4Cross = false
				t.ln4Target = t.config.ResolveIPv4DestAddress
				t.ln4TargetNetwork = "tcp4"
			}
		} else if t.config.Cross && t.config.ResolveIPv6DestAddress != nil {
			_ln4, err := net.ListenTCP("tcp4", t.config.ResolveIPv4SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp4 failed: %s", t.config.SrcPort, err.Error())
			}

			t.ln4 = _ln4
			t.ln4Proxy = false
			t.ln4Cross = true
			t.ln4Target = t.config.ResolveIPv6DestAddress
			t.ln4TargetNetwork = "tcp6"
		}
	} else {
		t.ln4 = nil
		t.ln4Proxy = false
		t.ln4Cross = false
		t.ln4Target = nil
		t.ln4TargetNetwork = ""
	}

	if ipcheck.SupportIPv6() {
		if t.config.ResolveIPv6DestAddress != nil {
			_ln6, err := net.ListenTCP("tcp6", t.config.ResolveIPv6SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp6 failed: %s", t.config.SrcPort, err.Error())
			}

			if t.config.IPv6SrcServerProxy.IsEnable(true) {
				t.ln6 = &proxyproto.Listener{
					Listener: _ln6,
				}
				t.ln6Proxy = true
				t.ln6Cross = false
				t.ln6Target = t.config.ResolveIPv6DestAddress
				t.ln6TargetNetwork = "tcp6"
			} else {
				t.ln6 = _ln6
				t.ln6Proxy = false
				t.ln6Cross = false
				t.ln6Target = t.config.ResolveIPv6DestAddress
				t.ln6TargetNetwork = "tcp6"
			}
		} else if t.config.Cross && t.config.ResolveIPv4DestAddress != nil {
			_ln6, err := net.ListenTCP("tcp6", t.config.ResolveIPv6SrcAddress)
			if err != nil {
				return fmt.Errorf("listen %d on tcp6 failed: %s", t.config.SrcPort, err.Error())
			}

			t.ln6 = _ln6
			t.ln6Proxy = false
			t.ln6Cross = true
			t.ln6Target = t.config.ResolveIPv4DestAddress
			t.ln6TargetNetwork = "tcp4"
		}
	} else {
		t.ln6 = nil
		t.ln6Proxy = false
		t.ln6Cross = false
		t.ln6Target = nil
		t.ln6TargetNetwork = ""
	}

	if t.ln4 == nil && t.ln6 == nil {
		return fmt.Errorf("no listen address")
	}

	if t.ln4Target == nil && t.ln6Target == nil {
		return fmt.Errorf("no target address")
	}

	t.stopchan = make(chan bool, 4)

	if t.ln4 != nil {
		go func() {
			defer func() {
				_ = t.ln4.Close()
				t.ln4 = nil
			}()

			logger.Infof("listen on %d (ipv4) start", t.config.SrcPort)
		MainCycle:
			for {
				select {
				case <-t.stopchan:
					break MainCycle
				default:
					// pass
				}

				status := t.accept(t.ln4,
					"tcp4",
					!t.ln4Cross && t.config.IPv4DestRequestProxy.IsEnable(true),
					t.config.IPv4DestRequestProxyVersion,
					t.ln4TargetNetwork,
					t.ln4Target)
				if status == StatusStop {
					break MainCycle
				}
			}

			logger.Infof("listen on %d (ipv4) stop", t.config.SrcPort)
		}()
	}

	if t.ln6 != nil {
		go func() {
			defer func() {
				_ = t.ln6.Close()
				t.ln6 = nil
			}()

			logger.Infof("listen on %d (ipv6) start", t.config.SrcPort)
		MainCycle:
			for {
				select {
				case <-t.stopchan:
					break MainCycle
				default:
					// pass
				}

				status := t.accept(t.ln6,
					"tcp6",
					!t.ln6Cross && t.config.IPv6DestRequestProxy.IsEnable(true),
					t.config.IPv6DestRequestProxyVersion,
					t.ln6TargetNetwork,
					t.ln6Target)
				if status == StatusStop {
					break MainCycle
				}
			}

			logger.Infof("listen on %d (ipv6) stop", t.config.SrcPort)
		}()
	}

	if !t.status.CompareAndSwap(StatusReady, StatusRunning) {
		return fmt.Errorf("server run failed: can not set status")
	}

	return nil
}

func (t *TcpServer) Stop() error {
	if t.ln4 == nil || !t.status.CompareAndSwap(StatusRunning, StatusStopping) {
		return nil
	}

	t.stopchan <- true
	t.stopchan <- true
	close(t.stopchan)
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
		t.allconn.Clear()
	}()

	t.swg.Wait()

	t.status.CompareAndSwap(StatusStopping, StatusFinished)
	return nil
}

func (t *TcpServer) forward(remoteAddr string, conn net.Conn, target net.Conn) {
	defer func() {
		r := recover()
		if r != nil {
			if err, ok := r.(error); ok {
				logger.Panicf("tcp forward panic error: %s", err.Error())
			} else {
				logger.Panicf("tcp forward panic error: %v", r)
			}
		}
	}()

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
		if err != nil && conn != nil && target != nil && t.status.Load() == StatusRunning {
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
		if err != nil && conn != nil && target != nil && t.status.Load() == StatusRunning {
			logger.Errorf("failed to forward from %s to %s: %v", target.RemoteAddr(), conn.RemoteAddr(), err)
		}
	}()

	<-stopchan
	return
}

func (t *TcpServer) accept(ln net.Listener, srcNetwork string, destProxy bool, destProxyVersion int, targetNetwork string, targetAddr *net.TCPAddr) string {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				logger.Panicf("listen on %d panic (error) : %s", t.config.SrcPort, err.Error())
			} else {
				logger.Panicf("listen on %d panic : %s", t.config.SrcPort, err.Error())
			}
		}
	}()

	if ln == nil {
		return StatusStop
	}

	conn, err := ln.Accept()
	if err != nil {
		logger.Errorf("listen on %d accecpt error: %s", t.config.SrcPort, err.Error())
		return StatusContinue
	}
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	if !t.controller.TcpNetworkAccept() {
		return StatusContinue
	}

	remoteAddr := conn.RemoteAddr()
	if remoteAddr == nil {
		return StatusContinue
	}

	remoteTCPAddr, err := net.ResolveTCPAddr(srcNetwork, remoteAddr.String())
	if err != nil {
		return StatusContinue
	}

	if !t.controller.RemoteAddrCheck(remoteTCPAddr) {
		return StatusContinue
	}

	target, err := net.DialTCP(targetNetwork, nil, targetAddr)
	if err != nil {
		logger.Errorf("Failed to connect to target %s: %v", targetAddr.String(), err)
		return StatusContinue
	}
	defer func() {
		if target != nil {
			_ = target.Close()
		}
	}()

	if destProxy {
		header := proxyproto.HeaderProxyFromAddrs(byte(destProxyVersion), remoteTCPAddr, targetAddr)
		_, err = header.WriteTo(target)
		if err != nil {
			logger.Errorf("Failed to write proxy header to target %s: %v", targetAddr.String(), err)
			return StatusContinue
		}
	}

	_conn := conn
	_target := target
	conn = nil
	target = nil
	go t.forward(remoteAddr.String(), _conn, _target)

	return StatusContinue
}
