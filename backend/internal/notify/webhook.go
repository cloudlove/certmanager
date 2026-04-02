package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// WebhookNotifier 自定义 Webhook 通知实现
type WebhookNotifier struct {
	config map[string]string
	client *http.Client
}

// NewWebhookNotifier 创建 Webhook 通知器
func NewWebhookNotifier(config map[string]string) *WebhookNotifier {
	return &WebhookNotifier{
		config: config,
		client: &http.Client{Timeout: 30},
	}
}

// Type 返回通知类型
func (w *WebhookNotifier) Type() string {
	return "webhook"
}

// WebhookMessage Webhook 消息结构
type WebhookMessage struct {
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Recipients []string `json:"recipients"`
	Timestamp  int64    `json:"timestamp"`
}

// Send 发送 Webhook 通知
func (w *WebhookNotifier) Send(ctx context.Context, msg NotifyMessage) error {
	webhookURL := w.config["webhook_url"]
	if webhookURL == "" {
		return fmt.Errorf("webhook_url not configured")
	}

	// 构建 Webhook 消息
	webhookMsg := WebhookMessage{
		Title:      msg.Title,
		Content:    msg.Content,
		Recipients: msg.Recipients,
		Timestamp:  0, // 可由调用方设置
	}

	jsonData, err := json.Marshal(webhookMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 添加自定义 Header（如果有配置）
	if authHeader := w.config["auth_header"]; authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	log.Printf("[Webhook Notification] Sent to %s, Title: %s", webhookURL, msg.Title)
	return nil
}
