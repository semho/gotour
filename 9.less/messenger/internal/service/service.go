package service

import (
	"github.com/google/uuid"
	"messenger/internal/config"
	"messenger/internal/model"
	"messenger/internal/storage"
)

type Services struct {
	User    UserService
	Message MessageService
	Chat    ChatService
}

func NewServices(store storage.Storage, cfg *config.Config) *Services {
	return &Services{
		User:    NewUserService(store, cfg.Storage.MaxUsers),
		Message: NewMessageService(store, cfg.Storage.MaxMessages),
		Chat:    NewChatService(store, cfg.Storage.MaxMessages),
	}
}

type UserService interface {
	CreateUser(username string) (*model.User, error)
	GetUser(requesterID, id uuid.UUID) (*model.User, error)
	GetAllUsers() ([]*model.User, error)
}

type MessageService interface {
	SendMessage(senderID, chatID uuid.UUID, text string) (*model.Message, error)
	GetMessage(id, requestingUserID uuid.UUID) (*model.Message, error)
	GetAllMessages() ([]*model.Message, error)
}

type ChatService interface {
	CreateChat(creatorID uuid.UUID, chatType model.ChatType, participantIDs []uuid.UUID) (
		*model.Chat,
		error,
	)
	GetChat(id uuid.UUID) (*model.Chat, error)
	AddUserToChat(chatID, userID, requesterID uuid.UUID) error
	RemoveUserFromChat(requesterID, chatID, userID uuid.UUID) error
	GetChatMessages(chatID, requesterID uuid.UUID) ([]*model.Message, error)
	GetAllChats() ([]*model.Chat, error)
}
