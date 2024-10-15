package handler

import (
	"messenger/internal/service"
)

type Handlers struct {
	User    *UserHandler
	Message *MessageHandler
	Chat    *ChatHandler
}

func NewHandlers(services *service.Services) *Handlers {
	return &Handlers{
		User:    NewUserHandler(services.User),
		Message: NewMessageHandler(services.Message),
		Chat:    NewChatHandler(services.Chat),
	}
}
