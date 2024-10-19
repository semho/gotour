package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"messenger/internal/handler"
	"messenger/internal/model"
)

type MockMessageService struct {
	mock.Mock
}

func (m *MockMessageService) SendMessage(senderID, chatID uuid.UUID, text string) (*model.Message, error) {
	args := m.Called(senderID, chatID, text)
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockMessageService) GetMessage(id, requestingUserID uuid.UUID) (*model.Message, error) {
	args := m.Called(id, requestingUserID)
	return args.Get(0).(*model.Message), args.Error(1)
}

func (m *MockMessageService) GetAllMessages() ([]*model.Message, error) {
	args := m.Called()
	return args.Get(0).([]*model.Message), args.Error(1)
}

func TestGetAllMessages(t *testing.T) {
	mockService := new(MockMessageService)
	messageHandler := handler.NewMessageHandler(mockService)

	messages := []*model.Message{
		{
			ID:        uuid.New(),
			SenderID:  uuid.New(),
			ChatID:    uuid.New(),
			Text:      "Привет",
			Timestamp: time.Now(),
			Status:    model.MessageStatusSent,
		},
		{
			ID:        uuid.New(),
			SenderID:  uuid.New(),
			ChatID:    uuid.New(),
			Text:      "Как дела?",
			Timestamp: time.Now(),
			Status:    model.MessageStatusDelivered,
		},
	}

	mockService.On("GetAllMessages").Return(messages, nil)

	_, _ = http.NewRequest("GET", "/messages", nil)
	rr := httptest.NewRecorder()

	messageHandler.GetAllMessages(rr)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []*model.Message
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, messages[0].ID, response[0].ID)
	assert.Equal(t, messages[1].ID, response[1].ID)

	mockService.AssertExpectations(t)
}

func TestSendMessage(t *testing.T) {
	mockService := new(MockMessageService)
	messageHandler := handler.NewMessageHandler(mockService)

	senderID := uuid.New()
	chatID := uuid.New()
	text := "Привет, это тестовое сообщение"

	message := &model.Message{
		ID:        uuid.New(),
		SenderID:  senderID,
		ChatID:    chatID,
		Text:      text,
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}

	mockService.On("SendMessage", senderID, chatID, text).Return(message, nil)

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": chatID,
			"text":   text,
		},
	)
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", senderID)
	req = req.WithContext(ctx)

	http.HandlerFunc(messageHandler.SendMessage).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.Message
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, message.ID, response.ID)
	assert.Equal(t, message.Text, response.Text)

	mockService.AssertExpectations(t)
}

func TestGetMessage(t *testing.T) {
	mockService := new(MockMessageService)
	messageHandler := handler.NewMessageHandler(mockService)

	receiverID := uuid.New()
	messageID := uuid.New()

	message := &model.Message{
		ID:        messageID,
		SenderID:  uuid.New(),
		ChatID:    uuid.New(),
		Text:      "Тестовое сообщение",
		Timestamp: time.Now(),
		Status:    model.MessageStatusRead,
	}

	mockService.On("GetMessage", messageID, receiverID).Return(message, nil)

	reqBody, _ := json.Marshal(
		map[string]any{
			"messageID": messageID,
		},
	)
	req, _ := http.NewRequest("POST", "/messages/read", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", receiverID)
	req = req.WithContext(ctx)

	http.HandlerFunc(messageHandler.GetMessage).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.Message
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, message.ID, response.ID)
	assert.Equal(t, message.Text, response.Text)

	mockService.AssertExpectations(t)
}

func TestSendMessageUnauthorized(t *testing.T) {
	mockService := new(MockMessageService)
	messageHandler := handler.NewMessageHandler(mockService)

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": uuid.New(),
			"text":   "Тестовое сообщение",
		},
	)
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	http.HandlerFunc(messageHandler.SendMessage).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGetMessageUnauthorized(t *testing.T) {
	mockService := new(MockMessageService)
	messageHandler := handler.NewMessageHandler(mockService)

	reqBody, _ := json.Marshal(
		map[string]any{
			"messageID": uuid.New(),
		},
	)
	req, _ := http.NewRequest("POST", "/messages/read", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	http.HandlerFunc(messageHandler.GetMessage).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestSendMessageError(t *testing.T) {
	mockService := new(MockMessageService)
	messageHandler := handler.NewMessageHandler(mockService)

	senderID := uuid.New()
	chatID := uuid.New()
	text := "Тестовое сообщение"

	mockService.On("SendMessage", senderID, chatID, text).Return(
		(*model.Message)(nil),
		errors.New("ошибка отправки сообщения"),
	)

	reqBody, _ := json.Marshal(
		map[string]any{
			"chatID": chatID,
			"text":   text,
		},
	)
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), "userID", senderID)
	req = req.WithContext(ctx)

	http.HandlerFunc(messageHandler.SendMessage).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
