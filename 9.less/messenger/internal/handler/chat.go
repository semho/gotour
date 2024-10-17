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
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case r.Method == http.MethodPost && path == "/chats":
		h.CreateChat(w, r)
	case r.Method == http.MethodGet && path == "/chats":
		h.GetAllChats(w)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/chats/id="):
		h.GetChat(w, r)
	case r.Method == http.MethodPost && strings.HasSuffix(path, "/chats/users"):
		h.AddUserToChat(w, r)
	case r.Method == http.MethodDelete && strings.HasSuffix(path, "/chats/users"):
		h.RemoveUserFromChat(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/chats/messages/id="):
		h.GetChatMessages(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *ChatHandler) GetAllChats(w http.ResponseWriter) {
	chats, err := h.chatService.GetAllChats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
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
	id := strings.TrimPrefix(r.URL.Path, "/chats/id=")
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
	var input struct {
		UserID uuid.UUID `json:"userID"`
		ChatID uuid.UUID `json:"chatID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.chatService.AddUserToChat(input.ChatID, input.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь успешно добавлен в чат"})
}

func (h *ChatHandler) RemoveUserFromChat(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID uuid.UUID `json:"userID"`
		ChatID uuid.UUID `json:"chatID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.chatService.RemoveUserFromChat(input.ChatID, input.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь успешно удален из чата"})
}

func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/chats/messages/id=")
	chatID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	var messages []*model.Message
	messages, err = h.chatService.GetChatMessages(chatID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
