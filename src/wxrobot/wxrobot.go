package wxrobot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"io"
	"net/http"
	"sync"
)

const (
	msgtypetext     = "text"
	msgtypemarkdown = "markdown"
)
const atall = "@all"

var WeChatRobotLock sync.Mutex

type WebhookText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type WebhookMarkdown struct {
	Content string `json:"content"`
}

type ReqWebhookMsg struct {
	MsgType  string           `json:"msgtype"`
	Text     *WebhookText     `json:"text,omitempty"`
	Markdown *WebhookMarkdown `json:"markdown,omitempty"`
}

type RespWebhookMsg struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func Send(msg string, atAll bool) error {
	if !config.IsReady() {
		panic("config is not ready")
	}

	webhook := config.GetConfig().API.Webhook

	if webhook == "" || msg == "" {
		return nil
	}

	if len([]byte(msg)) >= 2048 {
		return fmt.Errorf("msg too long")
	}

	data := ReqWebhookMsg{
		MsgType: msgtypetext,
		Text: &WebhookText{
			Content: msg,
		},
	}
	if atAll {
		data.Text.MentionedMobileList = []string{atall}
	}

	webhookData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json marshal error: %s", err.Error())
	}

	resp, err := http.Post(webhook, "application/json", bytes.NewBuffer(webhookData))
	if err != nil {
		return fmt.Errorf("http post error: %s", err.Error())
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body error: %s", err.Error())
	}

	var respWebhook RespWebhookMsg
	err = json.Unmarshal(respData, &respWebhook)
	if err != nil {
		return fmt.Errorf("json unmarshal response body error: %s", err.Error())
	}

	if respWebhook.ErrCode != 0 {
		return fmt.Errorf("send message error [code: %d]: %s", respWebhook.ErrCode, respWebhook.ErrMsg)
	}

	return nil
}
