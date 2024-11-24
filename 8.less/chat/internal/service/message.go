package service

import (
	"chat/internal/storage"
	kafka_v1 "chat/pkg/kafka/v1"
	"chat/pkg/logger"
	"context"
)

type MessageService interface {
	ProcessMessage(ctx context.Context, event *kafka_v1.ChatMessageEvent) error
}

type messageService struct {
	repo storage.MessageRepository
}

func NewMessageService(repo storage.MessageRepository) MessageService {
	return &messageService{
		repo: repo,
	}
}

func (s *messageService) ProcessMessage(ctx context.Context, event *kafka_v1.ChatMessageEvent) error {
	logger.Log.Info(
		"Processing message",
		"message_id", event.Payload.MessageId,
		"chat_id", event.Payload.ChatId,
	)

	// Проверяем и обновляем размер истории чата
	historySize, err := s.repo.GetChatHistorySize(ctx, event.Payload.ChatId)
	if err != nil {
		return err
	}

	currentSize, err := s.repo.GetCurrentMessagesCount(ctx, event.Payload.ChatId)
	if err != nil {
		return err
	}

	// Если превышен лимит, удаляем старые сообщения
	if historySize > 0 && currentSize >= historySize {
		if err := s.repo.DeleteOldestMessages(
			ctx,
			event.Payload.ChatId,
			currentSize-historySize+1,
		); err != nil {
			return err
		}
	}

	// Сохраняем новое сообщение
	//TODO: использовать kafka_v1.ChatMessageEvent_EVENT_TYPE_CREATED и другие события когда это потребуется, для этого надо создать доп методы под остальные события
	return s.repo.SaveMessage(ctx, event)
}
