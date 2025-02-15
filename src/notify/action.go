package notify

import (
	"github.com/SongZihuan/huan-springboard/src/smtpserver"
	"github.com/SongZihuan/huan-springboard/src/wxrobot"
	"runtime"
	"sync"
)

func SendStart() {
	go wxrobot.SendStart()
	go smtpserver.SendStart()
}

func SendWaitStop(reason string) {
	go wxrobot.SendWaitStop(reason)
	go smtpserver.SendWaitStop(reason)
}

func AsyncSendStop(exitcode int) {
	numGoroutine := runtime.NumGoroutine()

	go wxrobot.SendStop(exitcode, numGoroutine)
	go smtpserver.SendStop(exitcode, numGoroutine)
}

func SyncSendStop(exitcode int) {
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
	go wxrobot.SendTcpNotAccept()
	go smtpserver.SendTcpNotAccept()
}

func SendTcpStopAccept() {
	go wxrobot.SendTcpStopAccept()
	go smtpserver.SendTcpStopAccept()
}

func SendTcpReAccept() {
	go wxrobot.SendTcpReAccept()
	go smtpserver.SendTcpReAccept()
}

func SendSshBanned(ip string, to string, reason string) {
	go wxrobot.SendSshBanned(ip, to, reason)
	go smtpserver.SendSshBanned(ip, to, reason)
}

func SendSshSuccess(ip string, to string, mark string) {
	go wxrobot.SendSshSuccess(ip, to, mark)
	go smtpserver.SendSshSuccess(ip, to, mark)
}
