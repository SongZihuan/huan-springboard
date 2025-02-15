package smtpserver

import (
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/logger"
	"strings"
)

func printError(err error) {
	if err == nil {
		return
	}

	logger.Errorf("SMTP Send Error: %s", err.Error())
}

func SendStart() {
	printError(Send("服务启动完成", "服务启动/重启完成。"))
}

func SendWaitStop(reason string) {
	reason = strings.TrimSuffix(reason, "。")

	if reason == "" {
		reason = "无"
	}

	printError(Send("服务停止", fmt.Sprintf("服务即将停止（原因：%s）。", reason)))
}

func SendStop(exitcode int) {
	printError(Send("服务停止", fmt.Sprintf("服务停止。退出代码：%d。", exitcode)))
}

func SendTcpNotAccept() {
	printError(Send("网络高峰", fmt.Sprintf("网络高峰，Tcp服务暂停接收新请求。")))
}

func SendTcpStopAccept() {
	printError(Send("网络高峰", fmt.Sprintf("网络高峰，Tcp服务全部下线。")))
}

func SendTcpReAccept() {
	printError(Send("网络平稳", fmt.Sprintf("网络平稳，Tcp服务恢复。")))
}

func SendSshBanned(ip string, to string, reason string) {
	reason = strings.TrimSuffix(reason, "。")

	if reason == "" {
		reason = "无"
	}

	printError(Send("SSH异常请求（拒绝）", fmt.Sprintf("IP %s 连接到 %s 被拒（原因：%s）。", ip, to, reason)))
}

func SendSshSuccess(ip string, to string, mark string) {
	mark = strings.TrimSuffix(mark, "。")

	if mark == "" {
		mark = "无"
	}

	printError(Send("SSH请求（通过）", fmt.Sprintf("IP %s 连接到 %s 成功（备注：%s）。", ip, to, mark)))
}
