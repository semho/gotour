package storage

import (
	kafka_v1 "chat/pkg/kafka/v1"
	"context"
)

type MessageRepository interface {
	GetChatHistorySize(ctx context.Context, chatID string) (int, error)
	GetCurrentMessagesCount(ctx context.Context, chatID string) (int, error)
	DeleteOldestMessages(ctx context.Context, chatID string, count int) error
	SaveMessage(ctx context.Context, event *kafka_v1.ChatMessageEvent) error
}
