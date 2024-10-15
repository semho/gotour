package handler

import (
	"messenger/internal/service"
	"net/http"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: маршрутизация для сообщений
}
