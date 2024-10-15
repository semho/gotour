package service

import (
	"messenger/internal/storage/memory"
)

type Services struct {
	User    *UserService
	Message *MessageService
	Chat    *ChatService
}

func NewServices(db *memory.DB) *Services {
	return &Services{
		User:    NewUserService(db),
		Message: NewMessageService(db),
		Chat:    NewChatService(db),
	}
}
