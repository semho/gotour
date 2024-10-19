package service

import (
	"errors"
	"log"
	"messenger/internal/model"
	"messenger/internal/storage"
	"time"

	"github.com/google/uuid"
)

type messageService struct {
	storage     storage.Storage
	maxMessages int
}

func NewMessageService(storage storage.Storage, maxMessages int) MessageService {
	return &messageService{storage: storage, maxMessages: maxMessages}
}

func (s *messageService) SendMessage(senderID, chatID uuid.UUID, text string) (*model.Message, error) {
	if s.storage.GetMessageCount() >= s.maxMessages {
		return nil, errors.New("maximum number of messages reached")
	}

	chat, err := s.storage.GetChat(chatID)
	if err != nil {
		return nil, err
	}

	// являются ли отправитель и получатель участниками чата
	if !containsUser(chat.Participants, senderID) {
		return nil, errors.New("sender is not a participant of the chat")
	}

	switch chat.Type {
	case model.ChatTypePublic, model.ChatTypePrivate:
		// Все участники могут отправлять сообщения
	case model.ChatTypeReadOnly:
		// Только создатель может отправлять сообщения
		if senderID != chat.CreatorID {
			return nil, errors.New("only the creator can send messages in read-only chats")
		}
	}

	message := &model.Message{
		ID:        uuid.New(),
		SenderID:  senderID,
		ChatID:    chatID,
		Text:      text,
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}

	return s.storage.SendMessage(message)
}

func containsUser(participants []uuid.UUID, userID uuid.UUID) bool {
	for _, id := range participants {
		if id == userID {
			return true
		}
	}
	return false
}

func (s *messageService) GetMessage(id, requestingUserID uuid.UUID) (*model.Message, error) {
	message, err := s.storage.GetMessage(id)
	if err != nil {
		return nil, err
	}

	chat, err := s.storage.GetChat(message.ChatID)
	if err != nil {
		return nil, err
	}

	if chat.Type == model.ChatTypePrivate {
		if !containsUser(chat.Participants, requestingUserID) {
			return nil, errors.New("access denied")
		}
	}

	if message.SenderID != requestingUserID && message.Status == model.MessageStatusSent {
		message.Status = model.MessageStatusRead
		if err = s.storage.UpdateMessageStatus(id, model.MessageStatusRead); err != nil {
			log.Printf("Failed to update message status: %v", err)
		}
	}

	return message, nil
}

func (s *messageService) GetAllMessages() ([]*model.Message, error) {
	return s.storage.GetAllMessages()
}
