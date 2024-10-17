package service

import (
	"errors"
	"messenger/internal/model"
	"messenger/internal/storage"
	"time"

	"github.com/google/uuid"
)

type ChatService struct {
	storage  storage.Storage
	maxChats int
}

func NewChatService(storage storage.Storage, maxChats int) *ChatService {
	return &ChatService{storage: storage, maxChats: maxChats}
}

func (s *ChatService) CreateChat(chatType model.ChatType, participantIDs []uuid.UUID) (*model.Chat, error) {
	if s.storage.GetChatCount() >= s.maxChats {
		return nil, errors.New("maximum number of chats reached")
	}

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

func (s *ChatService) GetAllChats() ([]*model.Chat, error) {
	return s.storage.GetAllChats()
}
