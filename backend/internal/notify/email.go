package notify

import (
	"context"
	"fmt"
	"log"
)

// EmailNotifier 邮件通知实现
type EmailNotifier struct {
	config map[string]string
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(config map[string]string) *EmailNotifier {
	return &EmailNotifier{config: config}
}

// Type 返回通知类型
func (e *EmailNotifier) Type() string {
	return "email"
}

// Send 发送邮件通知 (模拟实现)
func (e *EmailNotifier) Send(ctx context.Context, msg NotifyMessage) error {
	// 模拟邮件发送，实际项目中应使用 SMTP
	log.Printf("[Email Notification] To: %v, Title: %s, Content: %s", msg.Recipients, msg.Title, msg.Content)

	// 这里可以添加真实的 SMTP 发送逻辑
	// smtpHost := e.config["smtp_host"]
	// smtpPort := e.config["smtp_port"]
	// username := e.config["username"]
	// password := e.config["password"]

	select {
	case <-ctx.Done():
		return fmt.Errorf("email send cancelled: %w", ctx.Err())
	default:
		return nil
	}
}
