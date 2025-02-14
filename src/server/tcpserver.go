package server

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/pires/go-proxyproto"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type TcpServer struct {
	status           atomic.Int32
	port             int64
	dest             string
	srcProxy         bool
	destProxy        bool
	destProxyVersion byte
	targetAddr       *net.TCPAddr
	listener         net.Listener
	swg              sync.WaitGroup
	allconn          sync.Map
	stopchan         chan bool
	controller       TcpController
}

type TcpServerOpt struct {
	port             int64
	dest             string
	srcProxy         bool
	destProxy        bool
	destProxyVersion int
	controller       TcpController
}

func NewTcpServer(opt *TcpServerOpt) (*TcpServer, error) {
	targetAddr, err := net.ResolveTCPAddr("tcp", opt.dest)
	if err != nil {
		return nil, err
	}

	res := &TcpServer{
		port:             opt.port,
		dest:             opt.dest,
		srcProxy:         opt.srcProxy,
		destProxy:        opt.destProxy,
		destProxyVersion: byte(opt.destProxyVersion),
		targetAddr:       targetAddr,
		controller:       opt.controller,
	}

	res.status.Store(StatusReady)

	return res, nil
}

func (t *TcpServer) Start() (err error) {
	if t.listener != nil || t.status.Load() != StatusReady {
		return nil
	}

	_listener, err := net.Listen("tcp", fmt.Sprintf(":%d", t.port))
	if err != nil {
		return fmt.Errorf("listen %d failed: %s", t.port, err.Error())
	}

	if t.srcProxy {
		t.listener = &proxyproto.Listener{
			Listener: _listener,
		}
	} else {
		t.listener = _listener
	}

	logger.Infof("listen on %d start", t.port)

	t.stopchan = make(chan bool, 2)

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

				if !t.controller.TcpNetworkAccept() {
					return StatusContinue
				}

				remoteAddr := conn.RemoteAddr()
				if remoteAddr == nil {
					return StatusContinue
				}

				remoteTCPAddr, err := net.ResolveTCPAddr("tcp", remoteAddr.String())
				if err != nil {
					return StatusContinue
				}

				if !t.controller.RemoteAddrCheck(remoteAddr.String()) {
					return StatusContinue
				}

				target, err := net.DialTCP("tcp", nil, t.targetAddr)
				if err != nil {
					logger.Errorf("Failed to connect to target %s: %v", t.dest, err)
					return StatusContinue
				}
				defer func() {
					if target != nil {
						_ = target.Close()
					}
				}()

				if t.destProxy {
					header := proxyproto.HeaderProxyFromAddrs(t.destProxyVersion, remoteTCPAddr, t.targetAddr)
					_, err = header.WriteTo(target)
					if err != nil {
						logger.Errorf("Failed to write proxy header to target %s: %v", t.dest, err)
						return StatusContinue
					}
				}

				_conn := conn
				_target := target
				conn = nil
				target = nil
				go t.forward(remoteAddr.String(), _conn, _target)

				return StatusContinue
			}()
			if status == StatusStop {
				break MainCycle
			}
		}

		logger.Infof("listen on %d stop", t.port)
	}()

	if !t.status.CompareAndSwap(StatusReady, StatusRunning) {
		return fmt.Errorf("server run failed: can not set status")
	}

	return nil
}

func (t *TcpServer) Stop() error {
	if t.listener == nil || !t.status.CompareAndSwap(StatusRunning, StatusStopping) {
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
		t.allconn.Clear()
	}()

	t.swg.Wait()

	t.status.CompareAndSwap(StatusStopping, StatusFinished)
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
		wg.Add(2)
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
