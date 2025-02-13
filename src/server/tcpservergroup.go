package server

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"sync"
)

var tcpServerGroupOnce sync.Once
var tcpServerGroup *TcpServerGroup

type TcpServerGroup struct {
	servers sync.Map
}

func NewTcpServerGroup() (res *TcpServerGroup) { // 单例模式
	tcpServerGroupOnce.Do(func() {
		tcpServerGroup = &TcpServerGroup{}
	})
	return tcpServerGroup
}

func (t *TcpServerGroup) StartAllServers() error {
	for _, f := range config.GetConfig().TCP.Forward {

		fmt.Println("TAG C", f.SrcPoint, f.DestAddress)
		server, err := NewTcpServer(f.SrcPoint, f.DestAddress)
		if err != nil {
			logger.Errorf("New TCP Server Error: %s\n", err)
		}

		_, loaded := t.servers.LoadOrStore(f.SrcPoint, server)
		if loaded {
			logger.Errorf("TCP Port Conflict: %d\n", f.SrcPoint)
		}

		err = server.Start()
		if err != nil {
			logger.Errorf("Start TCP Server Error: %s\n", err)
			// 没启动成功，但仍然保留在 Map 中，目的是提前发现可能的端口冲突（配置错误）
		}
	}

	return nil
}

func (t *TcpServerGroup) StopAllServers() error {
	var wg sync.WaitGroup

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

	wg.Wait()

	return nil
}
