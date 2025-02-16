package database

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StatusReady int32 = iota
	StatusRunning
	StatusStopping
	StatusFinished
)

type Cleaner struct {
	status   atomic.Int32
	stopchan chan bool
	timechan <-chan time.Time
	swg      sync.WaitGroup
}

var CleanerOnce sync.Once
var CleanerObj *Cleaner = nil

func NewCleaner() (*Cleaner, error) {
	CleanerOnce.Do(func() {
		if !config.IsReady() {
			panic("config is not ready")
		}

		obj := &Cleaner{
			stopchan: make(chan bool, 2),
			timechan: time.Tick(time.Duration(config.GetConfig().SQLite.Clean.ExecutionIntervalHour) * time.Hour),
		}

		obj.status.Store(StatusReady)

		CleanerObj = obj
	})

	return CleanerObj, nil
}

func (c *Cleaner) Start() error {
	if c.status.Load() != StatusReady {
		return nil
	}

	go func() {
		c.swg.Add(1)
		defer c.swg.Done()

		defer func() {
			r := recover()
			if r != nil {
				if err, ok := r.(error); ok {
					logger.Panicf("Database clean panic error: %s", err.Error())
				} else {
					logger.Panicf("Database clean panic: %v", err)
				}
			}
		}()

	MainCycle:
		for {
			c.clean()

			select {
			case <-c.stopchan:
				break MainCycle
			case <-c.timechan:
				//pass
			}
		}
	}()

	if !c.status.CompareAndSwap(StatusReady, StatusRunning) {
		return fmt.Errorf("status error")
	}

	return nil
}

func (c *Cleaner) clean() {
	go func() {
		c.swg.Add(1)
		defer c.swg.Done()

		defer func() {
			r := recover()
			if r != nil {
				if err, ok := r.(error); ok {
					logger.Panicf("Database clean iface record panic error: %s", err.Error())
				} else {
					logger.Panicf("Database clean iface record panic: %v", err)
				}
			}
		}()

		if config.GetConfig().SQLite.Clean.IfaceRecordSaveTime == -1 {
			logger.Errorf("skip clean iface record")
			return
		}

		logger.Infof("start clean iface record")
		err := CleanIfaceRecord(config.GetConfig().SQLite.Clean.IfaceRecordSaveTime)
		if err != nil {
			logger.Errorf("clean iface record error: %s", err.Error())
		}
	}()

	go func() {
		c.swg.Add(1)
		defer c.swg.Done()

		defer func() {
			r := recover()
			if r != nil {
				if err, ok := r.(error); ok {
					logger.Panicf("Database clean ssh connect record panic error: %s", err.Error())
				} else {
					logger.Panicf("Database clean ssh connect record panic: %v", err)
				}
			}
		}()

		if config.GetConfig().SQLite.Clean.SSHRecordSaveTime == -1 {
			logger.Errorf("skip clean ssh connect record")
			return
		}

		logger.Infof("start clean ssh connect record")
		err := CleanSshConnectRecord(config.GetConfig().SQLite.Clean.SSHRecordSaveTime)
		if err != nil {
			logger.Errorf("clean ssh connect record error: %s", err.Error())
		}
	}()
}

func (c *Cleaner) Stop() error {
	if !c.status.CompareAndSwap(StatusRunning, StatusStopping) {
		return nil
	}

	close(c.stopchan)

	c.swg.Wait()

	c.status.CompareAndSwap(StatusStopping, StatusFinished)
	return nil
}
