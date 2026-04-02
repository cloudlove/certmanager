package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// WeComNotifier 企业微信机器人通知实现
type WeComNotifier struct {
	config map[string]string
	client *http.Client
}

// NewWeComNotifier 创建企业微信通知器
func NewWeComNotifier(config map[string]string) *WeComNotifier {
	return &WeComNotifier{
		config: config,
		client: &http.Client{Timeout: 30},
	}
}

// Type 返回通知类型
func (w *WeComNotifier) Type() string {
	return "wecom"
}

// WeComMessage 企业微信消息结构
type WeComMessage struct {
	MsgType  string         `json:"msgtype"`
	Text     *WeComText     `json:"text,omitempty"`
	Markdown *WeComMarkdown `json:"markdown,omitempty"`
}

// WeComText 文本消息
type WeComText struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// WeComMarkdown Markdown消息
type WeComMarkdown struct {
	Content string `json:"content"`
}

// Send 发送企业微信通知
func (w *WeComNotifier) Send(ctx context.Context, msg NotifyMessage) error {
	webhookURL := w.config["webhook_url"]
	if webhookURL == "" {
		return fmt.Errorf("wecom webhook_url not configured")
	}

	// 构建企业微信消息
	wecomMsg := WeComMessage{
		MsgType: "markdown",
		Markdown: &WeComMarkdown{
			Content: fmt.Sprintf("**%s**\n\n%s", msg.Title, msg.Content),
		},
	}

	jsonData, err := json.Marshal(wecomMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal wecom message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create wecom request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send wecom message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("wecom webhook returned status: %d", resp.StatusCode)
	}

	log.Printf("[WeCom Notification] Sent to webhook, Title: %s", msg.Title)
	return nil
}
