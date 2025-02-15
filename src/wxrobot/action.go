package wxrobot

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"strings"
)

func printError(err error) {
	if err == nil {
		return
	}

	logger.Errorf("WxRobot Send Error: %s", err.Error())
}

func SendStart() {
	printError(Send("跳板机服务启动完成。", true))
}

func SendWaitStop(reason string) {
	reason = strings.TrimSuffix(reason, "。")

	if reason == "" {
		reason = "无"
	}

	printError(Send(fmt.Sprintf("跳板机服务即将停止（原因：%s）。", reason), true))
}

func SendStop(exitcode int) {
	printError(Send(fmt.Sprintf("跳板机服务停止。退出代码：%d。", exitcode), true))
}

func SendTcpNotAccept() {
	printError(Send(fmt.Sprintf("网络高峰，Tcp服务暂停接收新请求。"), true))
}

func SendTcpStopAccept() {
	printError(Send(fmt.Sprintf("网络高峰，Tcp服务全部下线。"), true))
}

func SendTcpReAccept() {
	printError(Send(fmt.Sprintf("网络平稳，Tcp服务恢复。"), true))
}

func SendSshBanned(ip string, to string, reason string) {
	reason = strings.TrimSuffix(reason, "。")

	if reason == "" {
		reason = "无"
	}

	printError(Send(fmt.Sprintf("IP %s 连接到 %s 被拒（原因：%s）。", ip, to, reason), true))
}

func SendSshSuccess(ip string, to string, mark string) {
	mark = strings.TrimSuffix(mark, "。")

	if mark == "" {
		mark = "无"
	}

	printError(Send(fmt.Sprintf("IP %s 连接到 %s 成功（备注：%s）。", ip, to, mark), false))
}
