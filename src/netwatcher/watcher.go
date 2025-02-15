package netwatcher

import (
	"errors"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/network"
	"github.com/shirou/gopsutil/v4/net"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StatusContinue = "continue"
	StatusStop     = "stop"
)

const (
	StatusReady int32 = iota
	StatusRunning
	StatusStopping
	StatusFinished
)

type NetWatcher struct {
	status    atomic.Int32
	ifaceName string
	iface     net.InterfaceStat
	stopchan  chan bool
	notices   sync.Map
}

type NotifyData struct {
	BytesSentPreSecond uint64
	BytesRecvPreSecond uint64
	SentLimit          uint64
	RecvLimit          uint64
	IsSentOK           bool
	IsRecvOK           bool
	IsOK               bool
	SpanOfSecond       float64
	UseRealLastRecord  bool
	IsStop             bool // stop 表示端口即将关闭
}

func NewNetWatcher() (*NetWatcher, error) {
	if config.GetConfig().TCP.RuleList.InterfaceName == "" {
		return nil, nil
	}

	iface, ok := network.Iface[config.GetConfig().TCP.RuleList.InterfaceName]
	if !ok {
		return nil, fmt.Errorf("interface %s not found", config.GetConfig().TCP.RuleList.InterfaceName)
	}

	res := &NetWatcher{
		ifaceName: iface.Name,
		iface:     iface,
		stopchan:  nil,
	}

	res.status.Store(StatusReady)
	return res, nil
}

func (t *NetWatcher) AddNotice(name string) chan *NotifyData {
	ch, _ := t.notices.LoadOrStore(name, make(chan *NotifyData, 15))
	return ch.(chan *NotifyData)
}

func (t *NetWatcher) Start() error {
	if t.stopchan != nil || t.status.Load() != StatusReady {
		return nil
	}

	t.stopchan = make(chan bool, 2)

	dlstopchan := make(chan bool, 2)
	ststopchan := make(chan bool, 2)
	clstopchan := make(chan bool, 2)

	go func() {
		defer func() {
			defer func() {
				_ = recover()
			}()

			dlstopchan <- true
		}()

		defer func() {
			defer func() {
				_ = recover()
			}()

			ststopchan <- true
		}()

		defer func() {
			defer func() {
				_ = recover()
			}()

			clstopchan <- true
		}()

		_stopchan := t.stopchan

		defer func() {
			defer func() {
				_ = recover()
			}()

			close(_stopchan)
		}()

		<-_stopchan
		t.stopchan = nil
	}()

	go func() {
		defer func() {
			defer func() {
				_ = recover()
			}()

			close(clstopchan)
		}()

	MainCycle:
		for {
			// 先sleep一段时间
			select {
			case <-time.After(time.Duration(config.GetConfig().TCP.RuleList.StatisticalPeriodSeconds) * 10 * time.Second):
				// pass
			case <-clstopchan:
				break MainCycle
			}

			status := func() string {
				var err error

				tmp1 := time.Duration(config.GetConfig().TCP.RuleList.StatisticalTimeSpanSeconds) * 3 * time.Second // 3倍span
				tmp2 := time.Hour * 24 * 31 * 3                                                                     // 三个月

				err = database.CleanIfaceRecord(t.ifaceName, max(tmp1, tmp2))
				if err != nil {
					logger.Errorf("Clean Interface data %s error: %s", t.ifaceName, err.Error())
					return StatusContinue
				}

				return StatusContinue
			}()
			if status == StatusContinue {
				// pass
			} else if status == StatusStop {
				break MainCycle
			}

			if status == StatusContinue {
				continue MainCycle
			}
		}
	}()

	go func() {
		defer func() {
			defer func() {
				_ = recover()
			}()

			close(dlstopchan)
		}()

	MainCycle:
		for {
			status := func() string {
				var err error

				data, err := t.getTargetInfo()
				if err != nil {
					logger.Errorf("Get Interface data %s error: %s", t.ifaceName, err.Error())
					return StatusContinue
				}

				err = database.AddIfaceRecord(t.ifaceName, data.BytesSent, data.BytesRecv, time.Now())
				if err != nil {
					logger.Errorf("Save Interface data to db %s error: %s", t.ifaceName, err.Error())
					return StatusContinue
				}

				return StatusContinue
			}()
			if status == StatusContinue {
				// pass
			} else if status == StatusStop {
				break MainCycle
			}

			select {
			case <-time.After(time.Duration(config.GetConfig().TCP.RuleList.DataCollectionCycleSeconds) * time.Second):
				// pass
			case <-dlstopchan:
				break MainCycle
			}

			if status == StatusContinue {
				continue MainCycle
			}
		}
	}()

	go func() {
		defer func() {
			defer func() {
				_ = recover()
			}()

			close(ststopchan)
		}()

	MainCycle:
		for {
			status := func() string {
				defer func() {
					r := recover()
					if r != nil {
						if err, ok := r.(error); ok {
							logger.Panicf("net watcher panic error: %s", err.Error())
						} else {
							logger.Panicf("net watcher panic: %v", r)
						}
					}
				}()

				var err error

				newRecord, err := database.FindIfaceNewRecord(t.ifaceName)
				if err != nil {
					if !errors.Is(err, database.ErrNotFound) {
						logger.Errorf("Get Interface %s record from db error: %s", t.ifaceName, err.Error())
					}
					return StatusContinue
				} else if time.Now().Sub(newRecord.Time) > 1*time.Minute {
					logger.Errorf("Get Interface %s record from db error: the time obtained is too far away from now ", t.ifaceName)
					return StatusContinue
				}

				lastDate := newRecord.Time.Add(-1 * time.Duration(config.GetConfig().TCP.RuleList.StatisticalTimeSpanSeconds) * time.Second)

				isRealLastRecord := true
				lastRecord, err := database.FindIfaceRecord(t.ifaceName, lastDate)
				if err != nil {
					if errors.Is(err, database.ErrNotFound) {
						_lastRecord, err := database.FindIfaceLastRecord(t.ifaceName)
						if err != nil {
							if !errors.Is(err, database.ErrNotFound) {
								logger.Errorf("Get Interface %s record from db error: %s", t.ifaceName, err.Error())
								return StatusContinue
							}
						}
						isRealLastRecord = false
						lastRecord = _lastRecord
					} else {
						logger.Errorf("Get Interface %s record from db error: %s", t.ifaceName, err.Error())
						return StatusContinue
					}
				} else if lastRecord.Time.After(lastDate) {
					logger.Errorf("last record time after than setting")
					return StatusContinue
				}

				if lastRecord != nil && newRecord != nil {
					bytesSent := newRecord.BytesSent - lastRecord.BytesSent
					bytesRecv := newRecord.BytesRecv - lastRecord.BytesRecv

					span := float64(newRecord.Time.Sub(lastRecord.Time)) / float64(time.Second)

					// 向上取整
					bytesSentPreSecond := uint64(math.Ceil(float64(bytesSent) / span))
					bytesRecvPreSecond := uint64(math.Ceil(float64(bytesRecv) / span))

					sentLimit := config.GetConfig().TCP.RuleList.SentLimit
					recvLimit := config.GetConfig().TCP.RuleList.RecvLimit

					logger.Tagf("sendLimit: %d; recvLimit: %d\n", sentLimit, recvLimit)

					isSentOK := sentLimit == 0 || bytesSentPreSecond <= sentLimit
					isRecvOK := recvLimit == 0 || bytesRecvPreSecond <= recvLimit

					isOK := isSentOK && isRecvOK

					data := &NotifyData{
						BytesSentPreSecond: bytesSentPreSecond,
						BytesRecvPreSecond: bytesRecvPreSecond,
						SentLimit:          sentLimit,
						RecvLimit:          recvLimit,
						IsSentOK:           isSentOK,
						IsRecvOK:           isRecvOK,
						IsOK:               isOK,
						SpanOfSecond:       span,
						UseRealLastRecord:  isRealLastRecord,
						IsStop:             false,
					}

					t.notices.Range(func(key, value any) bool {
						ch, ok := value.(chan *NotifyData)
						if !ok {
							return true
						}

						go func() {
							defer func() {
								_ = recover()
							}()
							ch <- data
						}()

						return true
					})

					if data.IsSentOK {
						logger.Debugf("%s 出方向【正常】：%s %s", t.ifaceName, t.networkSpeedBytesDisplay(bytesSentPreSecond), t.networkSpeedBitDisplay(bytesSentPreSecond*8))
					} else {
						logger.Debugf("%s 出方向【超过限制】：%s %s", t.ifaceName, t.networkSpeedBytesDisplay(bytesSentPreSecond), t.networkSpeedBitDisplay(bytesSentPreSecond*8))
					}

					if data.IsRecvOK {
						logger.Debugf("%s 入方向【正常】：%s %s", t.ifaceName, t.networkSpeedBytesDisplay(bytesRecvPreSecond), t.networkSpeedBitDisplay(bytesRecvPreSecond*8))
					} else {
						logger.Debugf("%s 入方向【超过限制】：%s %s", t.ifaceName, t.networkSpeedBytesDisplay(bytesRecvPreSecond), t.networkSpeedBitDisplay(bytesRecvPreSecond*8))
					}

					logger.Debugf("==== %s ====", time.Now().Format("2006-01-02 15:04:05"))
				}

				return StatusContinue
			}()
			if status == StatusContinue {
				// pass
			} else if status == StatusStop {
				break MainCycle
			}

			select {
			case <-time.After(time.Duration(config.GetConfig().TCP.RuleList.StatisticalPeriodSeconds) * time.Second):
				// pass
			case <-ststopchan:
				break MainCycle
			}

			if status == StatusContinue {
				continue MainCycle
			}
		}
	}()

	if !t.status.CompareAndSwap(StatusReady, StatusRunning) {
		return fmt.Errorf("start net watcher server failed: cann ot write status")
	}
	return nil
}

func (t *NetWatcher) Stop() {
	defer func() {
		// 防止报错：写入已关闭的channel
		_ = recover()
	}()

	if !t.status.CompareAndSwap(StatusRunning, StatusStopping) {
		return
	}

	if t.stopchan != nil {
		t.stopchan <- true
	}

	t.notices.Range(func(key, value any) bool {
		defer func() {
			_ = recover()
		}()

		ch, ok := value.(chan *NotifyData)
		if !ok {
			return true
		}

		ch <- &NotifyData{
			IsStop: true,
		}

		close(ch)

		return false
	})

	time.Sleep(time.Second * 1)
	_ = t.status.CompareAndSwap(StatusStopping, StatusFinished)
}

func (t *NetWatcher) getTargetInfo() (*net.IOCountersStat, error) {
	info, err := net.IOCounters(true) // pernic 为 true 表示分别返回信息
	if err != nil {
		return nil, err
	}

	for _, i := range info {
		if i.Name == t.ifaceName {
			return &i, nil
		}
	}

	return nil, fmt.Errorf("not found")
}

func (t *NetWatcher) networkSpeedBytesDisplay(bytesPreSecond uint64) string {
	defer func() {
		// 有除法，防止零除
		_ = recover()
	}()

	if bytesPreSecond == 0 {
		return "0B/S"
	} else if (bytesPreSecond / 1024) <= 0 {
		return fmt.Sprintf("%dB/S", bytesPreSecond)
	} else if (bytesPreSecond / 1024 / 1024) <= 0 {
		return fmt.Sprintf("%dKB/S", bytesPreSecond/1024)
	} else if (bytesPreSecond / 1024 / 1024 / 1024) <= 0 {
		return fmt.Sprintf("%dMB/S", bytesPreSecond/1024/1024)
	} else {
		return fmt.Sprintf("%dGB/S", bytesPreSecond/1024/1024/1024)
	}
}

func (t *NetWatcher) networkSpeedBitDisplay(bitPreSecond uint64) string {
	defer func() {
		// 有除法，防止零除
		_ = recover()
	}()

	if bitPreSecond == 0 {
		return "0bps"
	} else if (bitPreSecond / 1024) <= 0 {
		return fmt.Sprintf("%dbps", bitPreSecond)
	} else if (bitPreSecond / 1024 / 1024) <= 0 {
		return fmt.Sprintf("%dkbps", bitPreSecond/1024)
	} else if (bitPreSecond / 1024 / 1024 / 1024) <= 0 {
		return fmt.Sprintf("%dmbps", bitPreSecond/1024/1024)
	} else {
		return fmt.Sprintf("%dgbps", bitPreSecond/1024/1024/1024)
	}
}
