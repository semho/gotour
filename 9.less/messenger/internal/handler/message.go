package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"messenger/internal/service"
	"net/http"
	"strings"
)

type MessageHandler struct {
	messageService *service.MessageService
}

func NewMessageHandler(messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

func (h *MessageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	switch {
	case r.Method == http.MethodPost && path == "/messages":
		h.SendMessage(w, r)
	case r.Method == http.MethodGet && path == "/messages":
		h.GetAllMessages(w)
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/messages/read"):
		h.GetMessage(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *MessageHandler) GetAllMessages(w http.ResponseWriter) {
	messages, err := h.messageService.GetAllMessages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SenderID   uuid.UUID `json:"senderID"`
		ReceiverID uuid.UUID `json:"receiverID"`
		ChatID     uuid.UUID `json:"chatID"`
		Text       string    `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := h.messageService.SendMessage(input.SenderID, input.ReceiverID, input.ChatID, input.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MessageID  uuid.UUID `json:"messageID"`
		ReceiverID uuid.UUID `json:"receiverID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := h.messageService.GetMessage(input.MessageID, input.ReceiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}
