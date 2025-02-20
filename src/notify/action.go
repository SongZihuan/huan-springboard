package notify

import (
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/smtpserver"
	"github.com/SongZihuan/huan-springboard/src/wxrobot"
	"runtime"
	"sync"
)

var hasSendStart = false

func SendStart() {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendStart()
	go smtpserver.SendStart()

	hasSendStart = true
}

func SendWaitStop(reason string) {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendWaitStop(reason)
	go smtpserver.SendWaitStop(reason)
}

func AsyncSendStop(exitcode int) {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	if !hasSendStart {
		return
	}

	numGoroutine := runtime.NumGoroutine()

	go wxrobot.SendStop(exitcode, numGoroutine)
	go smtpserver.SendStop(exitcode, numGoroutine)
}

func SyncSendStop(exitcode int) {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	if !hasSendStart {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	numGoroutine := runtime.NumGoroutine()

	go func() {
		defer wg.Done()
		wxrobot.SendStop(exitcode, numGoroutine)
	}()

	go func() {
		defer wg.Done()
		smtpserver.SendStop(exitcode, numGoroutine)
	}()

	wg.Wait()
}

func SendTcpNotAccept() {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendTcpNotAccept()
	go smtpserver.SendTcpNotAccept()
}

func SendTcpStopAccept() {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendTcpStopAccept()
	go smtpserver.SendTcpStopAccept()
}

func SendTcpReAccept() {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendTcpReAccept()
	go smtpserver.SendTcpReAccept()
}

func SendSshBanned(ip string, to string, reason string) {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendSshBanned(ip, to, reason)
	go smtpserver.SendSshBanned(ip, to, reason)
}

func SendSshSuccess(ip string, to string, mark string) {
	if !config.IsReady() {
		panic("config is not ready")
	} else if config.GetConfig().Quite.IsEnable(false) {
		return
	}

	go wxrobot.SendSshSuccess(ip, to, mark)
	go smtpserver.SendSshSuccess(ip, to, mark)
}
