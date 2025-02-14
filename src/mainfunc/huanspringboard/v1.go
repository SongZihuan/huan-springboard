package huanspringboard

import (
	"errors"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/config/watcher"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/flagparser"
	"github.com/SongZihuan/huan-springboard/src/iface"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/redisserver"
	"github.com/SongZihuan/huan-springboard/src/server"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"os"
	"sync"
	"time"
)

func MainV1() int {
	var err error

	err = flagparser.InitFlag()
	if errors.Is(err, flagparser.StopFlag) {
		return 0
	} else if err != nil {
		return utils.ExitByError(err)
	}

	if !flagparser.IsReady() {
		return utils.ExitByErrorMsg("flag parser unknown error")
	}

	utils.SayHellof("%s", "The backend service program starts normally, thank you.")
	defer func() {
		utils.SayGoodByef("%s", "The backend service program is offline/shutdown normally, thank you.")
	}()

	cfgErr := config.InitConfig(flagparser.ConfigFile())
	if cfgErr != nil && cfgErr.IsError() {
		return utils.ExitByError(cfgErr)
	}

	if !config.IsReady() {
		return utils.ExitByErrorMsg("config parser unknown error")
	}

	err = logger.InitLogger(os.Stdout, os.Stderr)
	if err != nil {
		return utils.ExitByError(err)
	}

	if !logger.IsReady() {
		return utils.ExitByErrorMsg("logger unknown error")
	}

	if flagparser.RunAutoReload() {
		err = watcher.WatcherConfigFile()
		if err != nil {
			return utils.ExitByError(err)
		}
		defer watcher.CloseNotifyConfigFile()

		logger.Infof("Auto reload enable.")
	} else {
		logger.Infof("Auto reload disable.")
	}

	err = database.InitSQLite()
	if err != nil {
		fmt.Printf("init sqlite fail: %s\n", err.Error())
		return 1
	}
	defer database.CloseSQLite()

	err = redisserver.InitRedis()
	if err != nil {
		fmt.Printf("init redis fail: %s\n", err.Error())
		return 1
	}
	defer redisserver.CloseRedis()

	netWatcher, err := iface.NewNetWatcher()
	if err != nil {
		fmt.Printf("init iface fail: %s\n", err.Error())
		return 1
	}

	err = netWatcher.Start()
	if err != nil {
		fmt.Printf("start net watcher fail: %s\n", err.Error())
		return 1
	}
	defer netWatcher.Stop()

	tcpser := server.NewTcpServerGroup(netWatcher)

	logger.Executablef("%s", "ready")
	logger.Infof("run mode: %s", config.GetConfig().GlobalConfig.GetRunMode())

	err = tcpser.Start()
	if err != nil {
		fmt.Printf("start tcp server failed: %s\n", err.Error())
		return 1
	}
	defer func() {
		_ = tcpser.Stop()
	}()

	select {
	case <-config.GetSignalChan():
		func() { // 注意，此处不是协程
			var wg sync.WaitGroup
			go func() {
				wg.Add(1)
				defer wg.Done()

				netWatcher.Stop() // 提前关闭，同时代码上面的 defer 兜底
			}()

			go func() {
				wg.Add(1)
				defer wg.Done()

				_ = tcpser.Stop() // 提前关闭，同时代码上面的 defer 兜底
			}()

			wg.Wait()
		}() // 注意，此处不是协程

		time.Sleep(1 * time.Second)
		return 0
	}
	// 无法抵达
}
