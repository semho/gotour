package service

import (
	"errors"
	"log"
	"messenger/internal/model"
	"messenger/internal/storage"
	"time"

	"github.com/google/uuid"
)

type MessageService struct {
	storage     storage.Storage
	maxMessages int
}

func NewMessageService(storage storage.Storage, maxMessages int) *MessageService {
	return &MessageService{storage: storage, maxMessages: maxMessages}
}

func (s *MessageService) SendMessage(senderID, receiverID, chatID uuid.UUID, text string) (*model.Message, error) {
	if s.storage.GetMessageCount() >= s.maxMessages {
		return nil, errors.New("maximum number of messages reached")
	}

	chat, err := s.storage.GetChat(chatID)
	if err != nil {
		return nil, err
	}

	// являются ли отправитель и получатель участниками чата
	if !containsUser(chat.Participants, senderID) || !containsUser(chat.Participants, receiverID) {
		return nil, errors.New("sender or receiver is not a participant of the chat")
	}

	message := &model.Message{
		ID:         uuid.New(),
		SenderID:   senderID,
		ReceiverID: receiverID,
		ChatID:     chatID,
		Text:       text,
		Timestamp:  time.Now(),
		Status:     model.MessageStatusSent,
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

func (s *MessageService) GetMessage(id, requestingUserID uuid.UUID) (*model.Message, error) {
	message, err := s.storage.GetMessage(id)
	if err != nil {
		return nil, err
	}
	if message.SenderID != requestingUserID && message.ReceiverID != requestingUserID {
		return nil, errors.New("access denied")
	}

	if message.ReceiverID == requestingUserID && message.Status == model.MessageStatusSent {
		message.Status = model.MessageStatusRead
		if err = s.storage.UpdateMessageStatus(id, model.MessageStatusRead); err != nil {
			log.Printf("Failed to update message status: %v", err)
		}
	}

	return message, nil
}
