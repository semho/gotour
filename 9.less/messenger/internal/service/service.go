package service

import (
	"messenger/internal/config"
	"messenger/internal/storage"
)

type Services struct {
	User    *UserService
	Message *MessageService
	Chat    *ChatService
}

func NewServices(store storage.Storage, cfg *config.Config) *Services {
	return &Services{
		User:    NewUserService(store, cfg.Storage.MaxUsers),
		Message: NewMessageService(store, cfg.Storage.MaxMessages),
		Chat:    NewChatService(store, cfg.Storage.MaxMessages),
	}
}
