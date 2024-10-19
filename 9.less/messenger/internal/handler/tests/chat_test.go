package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"messenger/internal/handler"
	"messenger/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type MockChatService struct {
	mock.Mock
}

func (m *MockChatService) CreateChat(
	creatorID uuid.UUID,
	chatType model.ChatType,
	participantIDs []uuid.UUID,
) (*model.Chat, error) {
	args := m.Called(creatorID, chatType, participantIDs)
	return args.Get(0).(*model.Chat), args.Error(1)
}
func (m *MockChatService) GetChat(id uuid.UUID) (*model.Chat, error) {
	args := m.Called(id)
	return args.Get(0).(*model.Chat), args.Error(1)
}
func (m *MockChatService) AddUserToChat(chatID, userID, requesterID uuid.UUID) error {
	args := m.Called(chatID, userID, requesterID)
	return args.Error(0)
}
func (m *MockChatService) RemoveUserFromChat(requesterID, chatID, userID uuid.UUID) error {
	args := m.Called(requesterID, chatID, userID)
	return args.Error(0)
}
func (m *MockChatService) GetChatMessages(chatID, requesterID uuid.UUID) ([]*model.Message, error) {
	args := m.Called(chatID, requesterID)
	return args.Get(0).([]*model.Message), args.Error(1)
}
func (m *MockChatService) GetAllChats() ([]*model.Chat, error) {
	args := m.Called()
	return args.Get(0).([]*model.Chat), args.Error(1)
}

func TestCreateChat(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	creator := uuid.New()
	users := []uuid.UUID{uuid.New(), uuid.New()}

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         "public",
		Participants: users,
		CreatedAt:    time.Now(),
		CreatorID:    creator,
	}

	mockService.On("CreateChat", creator, model.ChatTypePublic, users).Return(chat, nil)

	reqBody, _ := json.Marshal(map[string]any{"type": "public", "participants": users})
	req, _ := http.NewRequest("POST", "/chats", bytes.NewBuffer(reqBody))
	req.Header.Set("User-ID", creator.String())
	rr := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userID", creator)
	req = req.WithContext(ctx)

	http.HandlerFunc(chatHandler.CreateChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.Chat
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, chat.ID, response.ID)
	assert.Equal(t, chat.Type, response.Type)

	mockService.AssertExpectations(t)
}

func TestGetAllChats(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	chats := []*model.Chat{
		{ID: uuid.New(), Type: model.ChatTypePublic},
		{ID: uuid.New(), Type: model.ChatTypePrivate},
	}

	mockService.On("GetAllChats").Return(chats, nil)

	_, _ = http.NewRequest("GET", "/chats", nil)
	rr := httptest.NewRecorder()
	chatHandler.GetAllChats(rr)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []*model.Chat
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, chats[0].ID, response[0].ID)
	assert.Equal(t, chats[1].ID, response[1].ID)

	mockService.AssertExpectations(t)
}

func TestGetChat(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	chatID := uuid.New()
	chat := &model.Chat{ID: chatID, Type: model.ChatTypePublic}

	mockService.On("GetChat", chatID).Return(chat, nil)

	req, _ := http.NewRequest("GET", "/chats/id="+chatID.String(), nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(chatHandler.GetChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.Chat
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, chat.ID, response.ID)
	assert.Equal(t, chat.Type, response.Type)

	mockService.AssertExpectations(t)
}

func TestAddUserToChat(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	requesterID := uuid.New()
	chatID := uuid.New()
	userID := uuid.New()

	mockService.On("AddUserToChat", chatID, userID, requesterID).Return(nil)

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": chatID,
			"userID": userID,
		},
	)
	req, _ := http.NewRequest("POST", "/chats/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", requesterID)
	req = req.WithContext(ctx)

	http.HandlerFunc(chatHandler.AddUserToChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Пользователь успешно добавлен в чат", response["message"])

	mockService.AssertExpectations(t)
}

func TestAddUserToChatError(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	requesterID := uuid.New()
	chatID := uuid.New()
	userID := uuid.New()

	errorMsg := "пользователь уже в чате"
	mockService.On("AddUserToChat", chatID, userID, requesterID).Return(errors.New(errorMsg))

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": chatID,
			"userID": userID,
		},
	)
	req, _ := http.NewRequest("POST", "/chats/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", requesterID)
	req = req.WithContext(ctx)

	http.HandlerFunc(chatHandler.AddUserToChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	responseBody := strings.TrimSpace(rr.Body.String())

	assert.Equal(t, errorMsg, responseBody)

	mockService.AssertExpectations(t)
}

func TestRemoveUserFromChat(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	requesterID := uuid.New()
	chatID := uuid.New()
	userID := uuid.New()

	mockService.On("RemoveUserFromChat", requesterID, chatID, userID).Return(nil)

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": chatID,
			"userID": userID,
		},
	)
	req, _ := http.NewRequest("DELETE", "/chats/users", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", requesterID)
	req = req.WithContext(ctx)

	http.HandlerFunc(chatHandler.RemoveUserFromChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Пользователь успешно удален из чата", response["message"])

	mockService.AssertExpectations(t)
}

func TestGetChatMessages(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	requesterID := uuid.New()
	chatID := uuid.New()
	now := time.Now()
	messages := []*model.Message{
		{
			ID:        uuid.New(),
			SenderID:  uuid.New(),
			ChatID:    chatID,
			Text:      "Hello",
			Timestamp: now,
			Status:    model.MessageStatusSent,
		},
		{
			ID:        uuid.New(),
			SenderID:  uuid.New(),
			ChatID:    chatID,
			Text:      "World",
			Timestamp: now.Add(time.Minute),
			Status:    model.MessageStatusDelivered,
		},
	}

	mockService.On("GetChatMessages", chatID, requesterID).Return(messages, nil)

	req, _ := http.NewRequest("GET", "/chats/messages/id="+chatID.String(), nil)
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", requesterID)
	req = req.WithContext(ctx)

	http.HandlerFunc(chatHandler.GetChatMessages).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []*model.Message
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)

	for i, msg := range response {
		assert.Equal(t, messages[i].ID, msg.ID)
		assert.Equal(t, messages[i].SenderID, msg.SenderID)
		assert.Equal(t, messages[i].ChatID, msg.ChatID)
		assert.Equal(t, messages[i].Text, msg.Text)
		assert.WithinDuration(t, messages[i].Timestamp, msg.Timestamp, time.Second)
		assert.Equal(t, messages[i].Status, msg.Status)
	}

	mockService.AssertExpectations(t)
}

func TestCreateChatUnauthorized(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	reqBody, _ := json.Marshal(
		map[string]any{
			"type":         "public",
			"participants": []uuid.UUID{uuid.New(), uuid.New()},
		},
	)
	req, _ := http.NewRequest("POST", "/chats", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	http.HandlerFunc(chatHandler.CreateChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAddUserToChatUnauthorized(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": uuid.New(),
			"userID": uuid.New(),
		},
	)
	req, _ := http.NewRequest("POST", "/chats/users", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	http.HandlerFunc(chatHandler.AddUserToChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetChatNotFound(t *testing.T) {
	mockService := new(MockChatService)
	chatHandler := handler.NewChatHandler(mockService)

	chatID := uuid.New()
	mockService.On("GetChat", chatID).Return((*model.Chat)(nil), errors.New("chat not found"))

	req, _ := http.NewRequest("GET", "/chats/id="+chatID.String(), nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(chatHandler.GetChat).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	mockService.AssertExpectations(t)
}
