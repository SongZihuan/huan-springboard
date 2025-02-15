package huanspringboard

import (
	"errors"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/config/watcher"
	"github.com/SongZihuan/huan-springboard/src/database"
	"github.com/SongZihuan/huan-springboard/src/flagparser"
	"github.com/SongZihuan/huan-springboard/src/ipcheck"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"github.com/SongZihuan/huan-springboard/src/netwatcher"
	"github.com/SongZihuan/huan-springboard/src/redisserver"
	"github.com/SongZihuan/huan-springboard/src/sshserver"
	"github.com/SongZihuan/huan-springboard/src/tcpserver"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"github.com/SongZihuan/huan-springboard/src/wxrobot"
	"os"
	"sync"
	"time"
)

func MainV1() (exitcode int) {
	var err error

	defer wxrobot.SendStop()

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
		if exitcode != 0 {
			utils.SayGoodByef("%s", "The backend service program is offline/shutdown with error, thank you.")
		} else {
			utils.SayGoodByef("%s", "The backend service program is offline/shutdown normally, thank you.")
		}
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

	if ipcheck.SupportIPv4() {
		logger.Infof("Server support ipv4.")
	} else {
		logger.Infof("Server dosen't support ipv4.")
	}

	if ipcheck.SupportIPv6() {
		logger.Infof("Server support ipv6.")
	} else {
		logger.Infof("Server dosen't support ipv6.")
	}

	if !ipcheck.SupportIPv4() && !ipcheck.SupportIPv6() {
		logger.Errorf("Server dosen't support ipv4 and ipv6.")
		return 1
	}

	err = database.InitSQLite()
	if err != nil {
		logger.Errorf("init sqlite fail: %s", err.Error())
		return 1
	}
	defer database.CloseSQLite()

	err = redisserver.InitRedis()
	if err != nil {
		logger.Errorf("init redis fail: %s\n", err.Error())
		return 1
	}
	defer redisserver.CloseRedis()

	netWatcher, err := netwatcher.NewNetWatcher()
	if err != nil {
		logger.Errorf("init net watcher fail: %s\n", err.Error())
		return 1
	}

	err = netWatcher.Start()
	if err != nil {
		logger.Errorf("start net watcher fail: %s\n", err.Error())
		return 1
	}
	defer netWatcher.Stop()

	tcpser := tcpserver.NewTcpServerGroup(netWatcher)
	sshser := sshserver.NewSshServerGroup()

	logger.Executablef("%s", "ready")
	logger.Infof("run mode: %s", config.GetConfig().GlobalConfig.GetRunMode())

	err = tcpser.Start()
	if err != nil {
		logger.Errorf("start tcp server failed: %s\n", err.Error())
		return 1
	}
	defer func() {
		_ = tcpser.Stop()
	}()

	err = sshser.Start()
	if err != nil {
		logger.Errorf("start ssh server failed: %s\n", err.Error())
		return 1
	}
	defer func() {
		_ = sshser.Stop()
	}()

	wxrobot.SendStart() // 此处是Start不是WaitStart

	select {
	case <-config.GetSignalChan():
		wxrobot.SendWaitStop("接收到退出信号")

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

			go func() {
				wg.Add(1)
				defer wg.Done()

				_ = sshser.Stop() // 提前关闭，同时代码上面的 defer 兜底
			}()

			wg.Wait()
		}() // 注意，此处不是协程

		time.Sleep(1 * time.Second)
		return 0
	}
	// 无法抵达
}
