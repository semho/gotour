package storage

import (
	"github.com/google/uuid"
	"messenger/internal/model"
)

type Storage interface {
	// User
	CreateUser(user *model.User) (*model.User, error)
	GetUserCount() int
	GetUser(id uuid.UUID) (*model.User, error)
	GetAllUsers() ([]*model.User, error)

	// Message
	GetMessageCount() int
	SendMessage(message *model.Message) (*model.Message, error)
	GetMessage(id uuid.UUID) (*model.Message, error)
	UpdateMessageStatus(id uuid.UUID, status model.MessageStatus) error
	GetAllMessages() ([]*model.Message, error)

	// Chat
	GetChatCount() int
	CreateChat(chat *model.Chat) (*model.Chat, error)
	GetChat(id uuid.UUID) (*model.Chat, error)
	AddUserToChat(chatID, userID uuid.UUID) error
	RemoveUserFromChat(chatID, userID uuid.UUID) error
	GetChatMessages(chatID uuid.UUID) ([]*model.Message, error)
	GetAllChats() ([]*model.Chat, error)
}
