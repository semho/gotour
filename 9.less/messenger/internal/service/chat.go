package service

import (
	"errors"
	"fmt"
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

func (s *ChatService) CreateChat(creatorID uuid.UUID, chatType model.ChatType, participantIDs []uuid.UUID) (
	*model.Chat,
	error,
) {
	if s.storage.GetChatCount() >= s.maxChats {
		return nil, errors.New("maximum number of chats reached")
	}

	if chatType != model.ChatTypePublic && chatType != model.ChatTypePrivate && chatType != model.ChatTypeReadOnly {
		return nil, errors.New("invalid chat type")
	}

	if !containsUser(participantIDs, creatorID) {
		participantIDs = append(participantIDs, creatorID)
	}

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         chatType,
		Participants: participantIDs,
		CreatedAt:    time.Now(),
		CreatorID:    creatorID,
	}
	return s.storage.CreateChat(chat)
}

func (s *ChatService) GetChat(id uuid.UUID) (*model.Chat, error) {
	return s.storage.GetChat(id)
}

func (s *ChatService) AddUserToChat(chatID, userID, requesterID uuid.UUID) error {
	chat, err := s.storage.GetChat(chatID)
	if err != nil {
		return err
	}

	_, err = s.storage.GetUser(requesterID)
	if err != nil {
		return fmt.Errorf("requester not found: %w", err)
	}

	switch chat.Type {
	case model.ChatTypePublic, model.ChatTypeReadOnly:
		// Любой может присоединиться
	case model.ChatTypePrivate:
		// Только создатель может добавлять пользователей
		if requesterID != chat.CreatorID {
			return errors.New("only the creator can add users to a private chat")
		}
	}

	return s.storage.AddUserToChat(chatID, userID)
}

func (s *ChatService) RemoveUserFromChat(requesterID, chatID, userID uuid.UUID) error {
	chat, err := s.storage.GetChat(chatID)
	if err != nil {
		return fmt.Errorf("failed to get chat: %w", err)
	}

	isRequesterInChat, err := s.storage.IsUserInChat(chatID, requesterID)
	if err != nil {
		return fmt.Errorf("failed to check if requester is in chat: %w", err)
	}
	if !isRequesterInChat {
		return errors.New("requester is not a member of the chat")
	}

	isUserInChat, err := s.storage.IsUserInChat(chatID, userID)
	if err != nil {
		return fmt.Errorf("failed to check if user is in chat: %w", err)
	}
	if !isUserInChat {
		return errors.New("user is not a member of the chat")
	}

	if requesterID != chat.CreatorID && userID == requesterID {
		return errors.New("you cannot delete the creator")
	}
	if requesterID != chat.CreatorID && userID != requesterID {
		return errors.New("you cannot delete other members")
	}

	isRequesterCreator := requesterID == chat.CreatorID
	isUserCreator := userID == chat.CreatorID
	isDeletingSelf := userID == requesterID

	switch {
	case isRequesterCreator:
		// Создатель пытается удалить себя
		if isDeletingSelf {
			return errors.New("chat creator cannot remove themselves")
		}
		// Создатель может удалить любого другого пользователя
	case !isRequesterCreator:
		// Не создатель пытается удалить создателя
		if isUserCreator {
			return errors.New("non-creator cannot remove the chat creator")
		}
		// Не создатель пытается удалить другого пользователя
		if !isDeletingSelf {
			return errors.New("non-creator can only remove themselves")
		}
	}

	if err = s.storage.RemoveUserFromChat(chatID, userID); err != nil {
		return fmt.Errorf("failed to remove user from chat: %w", err)
	}

	return nil
}

func (s *ChatService) GetChatMessages(chatID, requesterID uuid.UUID) ([]*model.Message, error) {
	chat, err := s.storage.GetChat(chatID)
	if err != nil {
		return nil, err
	}

	_, err = s.storage.GetUser(requesterID)
	if err != nil {
		return nil, fmt.Errorf("requester not found: %w", err)
	}

	switch chat.Type {
	case model.ChatTypePublic, model.ChatTypeReadOnly:
		// Любой пользователь может читать сообщения
	case model.ChatTypePrivate:
		// Только участники могут читать сообщения
		if !containsUser(chat.Participants, requesterID) {
			return nil, errors.New("user is not a participant of this private chat")
		}
	}

	return s.storage.GetChatMessages(chatID)
}

func (s *ChatService) GetAllChats() ([]*model.Chat, error) {
	return s.storage.GetAllChats()
}
