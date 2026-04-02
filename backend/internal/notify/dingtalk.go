package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// DingTalkNotifier 钉钉机器人通知实现
type DingTalkNotifier struct {
	config map[string]string
	client *http.Client
}

// NewDingTalkNotifier 创建钉钉通知器
func NewDingTalkNotifier(config map[string]string) *DingTalkNotifier {
	return &DingTalkNotifier{
		config: config,
		client: &http.Client{Timeout: 30},
	}
}

// Type 返回通知类型
func (d *DingTalkNotifier) Type() string {
	return "dingtalk"
}

// DingTalkMessage 钉钉消息结构
type DingTalkMessage struct {
	MsgType  string            `json:"msgtype"`
	Text     *DingTalkText     `json:"text,omitempty"`
	Markdown *DingTalkMarkdown `json:"markdown,omitempty"`
}

// DingTalkText 文本消息
type DingTalkText struct {
	Content string `json:"content"`
}

// DingTalkMarkdown Markdown消息
type DingTalkMarkdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Send 发送钉钉通知
func (d *DingTalkNotifier) Send(ctx context.Context, msg NotifyMessage) error {
	webhookURL := d.config["webhook_url"]
	if webhookURL == "" {
		return fmt.Errorf("dingtalk webhook_url not configured")
	}

	// 构建钉钉消息
	dingMsg := DingTalkMessage{
		MsgType: "markdown",
		Markdown: &DingTalkMarkdown{
			Title: msg.Title,
			Text:  fmt.Sprintf("## %s\n\n%s", msg.Title, msg.Content),
		},
	}

	jsonData, err := json.Marshal(dingMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal dingtalk message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create dingtalk request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send dingtalk message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk webhook returned status: %d", resp.StatusCode)
	}

	log.Printf("[DingTalk Notification] Sent to webhook, Title: %s", msg.Title)
	return nil
}
