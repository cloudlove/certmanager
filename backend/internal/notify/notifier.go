package notify

import (
	"context"
	"fmt"
)

// NotifyMessage 通知消息结构体
type NotifyMessage struct {
	Title      string
	Content    string
	Recipients []string
}

// Notifier 通知渠道接口
type Notifier interface {
	Send(ctx context.Context, msg NotifyMessage) error
	Type() string
}

// NewNotifier 创建通知渠道实例
func NewNotifier(channelType string, config map[string]string) (Notifier, error) {
	switch channelType {
	case "email":
		return NewEmailNotifier(config), nil
	case "dingtalk":
		return NewDingTalkNotifier(config), nil
	case "wecom":
		return NewWeComNotifier(config), nil
	case "webhook":
		return NewWebhookNotifier(config), nil
	default:
		return nil, fmt.Errorf("unsupported notification channel: %s", channelType)
	}
}
