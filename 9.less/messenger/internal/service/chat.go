package service

import (
	"messenger/internal/model"
	"messenger/internal/storage/memory"
	"time"

	"github.com/google/uuid"
)

type ChatService struct {
	storage *memory.DB
}

func NewChatService(storage *memory.DB) *ChatService {
	return &ChatService{storage: storage}
}

func (s *ChatService) CreateChat(chatType model.ChatType, participantIDs []uuid.UUID) (*model.Chat, error) {
	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         chatType,
		Participants: participantIDs,
		CreatedAt:    time.Now(),
	}
	return s.storage.CreateChat(chat)
}

func (s *ChatService) GetChat(id uuid.UUID) (*model.Chat, error) {
	return s.storage.GetChat(id)
}

func (s *ChatService) AddUserToChat(chatID, userID uuid.UUID) error {
	return s.storage.AddUserToChat(chatID, userID)
}

func (s *ChatService) RemoveUserFromChat(chatID, userID uuid.UUID) error {
	return s.storage.RemoveUserFromChat(chatID, userID)
}

func (s *ChatService) GetChatMessages(chatID uuid.UUID) ([]*model.Message, error) {
	return s.storage.GetChatMessages(chatID)
}
