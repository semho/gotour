package handler

import (
	"encoding/json"
	"github.com/google/uuid"
	"messenger/internal/model"
	"messenger/internal/service"
	"net/http"
	"strings"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/chats":
		h.CreateChat(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/chats/"):
		h.GetChat(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/users"):
		h.AddUserToChat(w, r)
	case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/users/"):
		h.RemoveUserFromChat(w, r)
	case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/messages"):
		h.GetChatMessages(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type         model.ChatType `json:"type"`
		Participants []uuid.UUID    `json:"participants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	chat, err := h.chatService.CreateChat(input.Type, input.Participants)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

func (h *ChatHandler) GetChat(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/chats/")
	chatID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	chat, err := h.chatService.GetChat(chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chat)
}

func (h *ChatHandler) AddUserToChat(w http.ResponseWriter, r *http.Request) {
	// TODO: добавление пользователя в чат
}

func (h *ChatHandler) RemoveUserFromChat(w http.ResponseWriter, r *http.Request) {
	// TODO: удаление пользователя из чата
}

func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	// TODO: получение сообщений чата
}
