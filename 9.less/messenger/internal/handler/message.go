package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"messenger/internal/middleware"
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

func (h *MessageHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("/messages", h.handleMessages())
	mux.Handle("/messages/", h.handleMessages())
}

func (h *MessageHandler) handleMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")

		switch {
		case r.Method == http.MethodGet && path == "/messages":
			h.GetAllMessages(w)
		default:
			// применяем middleware.Auth
			middleware.Auth(
				h.authenticatedRequests(),
			).ServeHTTP(w, r)
		}
	}
}

func (h *MessageHandler) authenticatedRequests() func(
	w http.ResponseWriter,
	r *http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSuffix(r.URL.Path, "/")
		switch {
		case r.Method == http.MethodPost && path == "/messages":
			h.SendMessage(w, r)
		case r.Method == http.MethodPost && strings.HasPrefix(path, "/messages/read"):
			h.GetMessage(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
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
	senderID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var input struct {
		ChatID uuid.UUID `json:"chatID"`
		Text   string    `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := h.messageService.SendMessage(senderID, input.ChatID, input.Text)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	receiverID, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		MessageID uuid.UUID `json:"messageID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	message, err := h.messageService.GetMessage(input.MessageID, receiverID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}
