package service

import (
	"bytes"
	"chat/internal/middleware"
	"chat/internal/models"
	pb "chat/pkg/chat/v1"
	kafka_v1 "chat/pkg/kafka/v1"
	"chat/pkg/logger"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MockProducer struct {
	mock.Mock
}

func (m *MockProducer) SendMessage(ctx context.Context, event *kafka_v1.ChatMessageEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorage) CreateSession(ctx context.Context, session *models.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

// MockStorage - мок для хранилища
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) SaveAnonNickname(ctx context.Context, chatID, sessionID, nickname string) error {
	args := m.Called(ctx, chatID, sessionID, nickname)
	return args.Error(0)
}

func (m *MockStorage) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockStorage) CreateChat(ctx context.Context, chat *models.Chat) error {
	args := m.Called(ctx, chat)
	return args.Error(0)
}

func (m *MockStorage) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).(*models.Chat), args.Error(1)
}

func (m *MockStorage) DeleteChat(ctx context.Context, chatID string) error {
	args := m.Called(ctx, chatID)
	return args.Error(0)
}

func (m *MockStorage) SetChatTTL(ctx context.Context, chatID string, ttl time.Time) error {
	args := m.Called(ctx, chatID, ttl)
	return args.Error(0)
}

func (m *MockStorage) AddMessage(ctx context.Context, message *models.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockStorage) GetChatHistory(ctx context.Context, chatID string) ([]*models.Message, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]*models.Message), args.Error(1)
}

func (m *MockStorage) RequestChatAccess(ctx context.Context, chatID, sessionID string) error {
	args := m.Called(ctx, chatID, sessionID)
	return args.Error(0)
}

func (m *MockStorage) GetAccessRequests(ctx context.Context, chatID string) ([]string, error) {
	args := m.Called(ctx, chatID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) GrantChatAccess(ctx context.Context, chatID, sessionID string) error {
	args := m.Called(ctx, chatID, sessionID)
	return args.Error(0)
}

func (m *MockStorage) IsChatOwner(ctx context.Context, chatID, sessionID string) (bool, error) {
	args := m.Called(ctx, chatID, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) HasChatAccess(ctx context.Context, chatID, sessionID string) (bool, error) {
	args := m.Called(ctx, chatID, sessionID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) GetAndIncrementAnonCount(ctx context.Context, chatID string) (int, error) {
	args := m.Called(ctx, chatID)
	return args.Int(0), args.Error(1)
}

func (m *MockStorage) GetDefaultHistorySize() int {
	args := m.Called()
	return args.Int(0)
}

func createContextWithSession(sessionID string) context.Context {
	md := metadata.New(map[string]string{"session_id": sessionID})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	return context.WithValue(ctx, middleware.SessionIDKey, sessionID)
}

func TestMain(m *testing.M) {
	logger.Init()

	var buf bytes.Buffer
	oldOutput := logger.Log.Handler().(*logger.ColorHandler).W
	logger.Log.Handler().(*logger.ColorHandler).W = &buf
	defer func() {
		logger.Log.Handler().(*logger.ColorHandler).W = oldOutput
	}()

	// Запуск тестов
	exitCode := m.Run()

	if testing.Verbose() {
		_, err := io.Copy(os.Stdout, &buf)
		if err != nil {
			return
		}
	}

	os.Exit(exitCode)
}

func TestCreateSession(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	ctx := context.Background()
	req := &pb.CreateSessionRequest{Nickname: "testuser"}

	mockStorage.On("CreateSession", mock.Anything, mock.AnythingOfType("*models.Session")).Return(nil)

	resp, err := service.CreateSession(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.Nickname, resp.Nickname)
	assert.NotEmpty(t, resp.Id)

	mockStorage.AssertExpectations(t)
}

func TestCreateChat(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.CreateChatRequest{
		HistorySize: 100,
		TtlSeconds:  3600,
		ReadOnly:    false,
		Private:     true,
	}

	mockSession := &models.Session{
		ID:       sessionID,
		Nickname: "TestUser",
	}

	mockStorage.On("GetSession", mock.Anything, sessionID).Return(mockSession, nil)
	mockStorage.On("CreateChat", mock.Anything, mock.AnythingOfType("*models.Chat")).Return(nil)

	resp, err := service.CreateChat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(100), resp.HistorySize)
	assert.Equal(t, true, resp.Private)
	assert.Equal(t, false, resp.ReadOnly)
	assert.NotEmpty(t, resp.Id)
	assert.Equal(t, sessionID, resp.OwnerId)

	mockStorage.AssertExpectations(t)
}

func TestCreateChat_DefaultHistorySize(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.CreateChatRequest{
		HistorySize: 0,
		TtlSeconds:  3600,
		ReadOnly:    false,
		Private:     true,
	}

	mockSession := &models.Session{
		ID:       sessionID,
		Nickname: "TestUser",
	}

	defaultHistorySize := 1000
	mockStorage.On("GetSession", mock.Anything, sessionID).Return(mockSession, nil)
	mockStorage.On("CreateChat", mock.Anything, mock.AnythingOfType("*models.Chat")).Return(nil)
	mockStorage.On("GetDefaultHistorySize").Return(defaultHistorySize)

	resp, err := service.CreateChat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, int32(defaultHistorySize), resp.HistorySize)
	assert.Equal(t, true, resp.Private)
	assert.Equal(t, false, resp.ReadOnly)
	assert.NotEmpty(t, resp.Id)
	assert.Equal(t, sessionID, resp.OwnerId)

	mockStorage.AssertExpectations(t)
}

func TestDeleteChat(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.DeleteChatRequest{ChatId: "test_chat_id"}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(true, nil)
	mockStorage.On("DeleteChat", mock.Anything, req.ChatId).Return(nil)

	resp, err := service.DeleteChat(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.IsType(t, &emptypb.Empty{}, resp)

	mockStorage.AssertExpectations(t)
}

func TestDeleteChat_NotOwner(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.DeleteChatRequest{ChatId: "test_chat_id"}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(false, nil)

	resp, err := service.DeleteChat(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))

	mockStorage.AssertExpectations(t)
}

func TestSetChatTTL(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.SetChatTTLRequest{
		ChatId:     "test_chat_id",
		TtlSeconds: 3600,
	}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(true, nil)
	mockStorage.On("SetChatTTL", mock.Anything, req.ChatId, mock.AnythingOfType("time.Time")).Return(nil)

	resp, err := service.SetChatTTL(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.IsType(t, &emptypb.Empty{}, resp)

	mockStorage.AssertExpectations(t)
}

func TestSetChatTTL_NotOwner(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.SetChatTTLRequest{
		ChatId:     "test_chat_id",
		TtlSeconds: 3600,
	}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(false, nil)

	resp, err := service.SetChatTTL(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))

	mockStorage.AssertExpectations(t)
}

func TestSendMessage(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.SendMessageRequest{
		ChatId: "test_chat_id",
		Text:   "Hello, world!",
	}

	mockChat := &models.Chat{
		ID:       "test_chat_id",
		OwnerID:  "owner_id",
		ReadOnly: false,
		Private:  false,
	}

	mockSession := &models.Session{
		ID:            sessionID,
		Nickname:      "TestUser",
		AnonNicknames: make(map[string]string),
	}

	mockStorage.On("GetChat", mock.Anything, req.ChatId).Return(mockChat, nil)
	mockStorage.On("GetSession", mock.Anything, sessionID).Return(mockSession, nil)

	mockProducer.On(
		"SendMessage",
		mock.Anything,
		mock.MatchedBy(
			func(event *kafka_v1.ChatMessageEvent) bool {
				return event.Payload.ChatId == req.ChatId &&
					event.Payload.Text == req.Text &&
					event.Payload.SessionId == sessionID &&
					event.Payload.Nickname == mockSession.Nickname &&
					event.Metadata.EventType == kafka_v1.ChatMessageEvent_EVENT_TYPE_CREATED
			},
		),
	).Return(nil)

	resp, err := service.SendMessage(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.ChatId, resp.ChatId)
	assert.Equal(t, req.Text, resp.Text)
	assert.Equal(t, sessionID, resp.SessionId)
	assert.Equal(t, "TestUser", resp.Nickname)

	mockStorage.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestSendMessage_AnonymousUser(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.SendMessageRequest{
		ChatId: "test_chat_id",
		Text:   "Hello, world!",
	}

	mockChat := &models.Chat{
		ID:       "test_chat_id",
		OwnerID:  "owner_id",
		ReadOnly: false,
		Private:  false,
	}

	mockSession := &models.Session{
		ID:            sessionID,
		Nickname:      "", // Пустой никнейм для анонимного пользователя
		AnonNicknames: make(map[string]string),
	}

	mockStorage.On("GetChat", mock.Anything, req.ChatId).Return(mockChat, nil)
	mockStorage.On("GetSession", mock.Anything, sessionID).Return(mockSession, nil)
	mockStorage.On("GetAndIncrementAnonCount", mock.Anything, req.ChatId).Return(1, nil)
	mockStorage.On("SaveAnonNickname", mock.Anything, req.ChatId, sessionID, "Аноним #1").Return(nil)

	mockProducer.On(
		"SendMessage",
		mock.AnythingOfType("*context.valueCtx"),
		mock.MatchedBy(
			func(event *kafka_v1.ChatMessageEvent) bool {
				return event != nil &&
					event.Payload != nil &&
					event.Payload.ChatId == req.ChatId &&
					event.Payload.Text == req.Text &&
					event.Payload.SessionId == sessionID &&
					event.Payload.Nickname == "Аноним #1"
			},
		),
	).Return(nil)

	resp, err := service.SendMessage(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, req.ChatId, resp.ChatId)
	assert.Equal(t, req.Text, resp.Text)
	assert.Equal(t, sessionID, resp.SessionId)
	assert.Equal(t, "Аноним #1", resp.Nickname)

	mockStorage.AssertExpectations(t)
	mockProducer.AssertExpectations(t)
}

func TestGetChatHistory(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	chatID := "test_chat_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.GetChatHistoryRequest{ChatId: chatID}

	mockChat := &models.Chat{
		ID:       chatID,
		OwnerID:  "owner_id",
		ReadOnly: false,
		Private:  false,
	}

	mockSession := &models.Session{
		ID:       sessionID,
		Nickname: "TestUser",
	}

	mockMessages := []*models.Message{
		{ID: "1", ChatID: chatID, SessionID: "session1", Text: "Hello", Timestamp: time.Now()},
		{ID: "2", ChatID: chatID, SessionID: "session2", Text: "Hi", Timestamp: time.Now()},
	}

	mockStorage.On("GetSession", mock.Anything, sessionID).Return(mockSession, nil)
	mockStorage.On("GetChat", mock.Anything, chatID).Return(mockChat, nil)
	mockStorage.On("GetChatHistory", mock.Anything, chatID).Return(mockMessages, nil)

	resp, err := service.GetChatHistory(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Messages, 2)
	assert.Equal(t, mockMessages[0].Text, resp.Messages[0].Text)
	assert.Equal(t, mockMessages[1].Text, resp.Messages[1].Text)

	mockStorage.AssertExpectations(t)
}

func TestRequestChatAccess(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.RequestChatAccessRequest{
		ChatId: "test_chat_id",
	}

	mockChat := &models.Chat{
		ID:      req.ChatId,
		Private: true,
	}

	mockStorage.On("GetChat", mock.Anything, req.ChatId).Return(mockChat, nil)
	mockStorage.On("HasChatAccess", mock.Anything, req.ChatId, sessionID).Return(false, nil)
	mockStorage.On("RequestChatAccess", mock.Anything, req.ChatId, sessionID).Return(nil)

	resp, err := service.RequestChatAccess(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "request_sent", resp.Status)

	mockStorage.AssertExpectations(t)
}

func TestGetAccessRequests(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.GetAccessRequestsRequest{ChatId: "test_chat_id"}

	mockSessionIDs := []string{"session1", "session2"}
	mockSessions := []*models.Session{
		{ID: "session1", Nickname: "user1"},
		{ID: "session2", Nickname: "user2"},
	}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(true, nil)
	mockStorage.On("GetAccessRequests", mock.Anything, req.ChatId).Return(mockSessionIDs, nil)
	mockStorage.On("GetSession", mock.Anything, "session1").Return(mockSessions[0], nil)
	mockStorage.On("GetSession", mock.Anything, "session2").Return(mockSessions[1], nil)

	resp, err := service.GetAccessRequests(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Requests, 2)
	assert.Equal(t, mockSessions[0].ID, resp.Requests[0].Id)
	assert.Equal(t, mockSessions[1].ID, resp.Requests[1].Id)

	mockStorage.AssertExpectations(t)
}

func TestGetAccessRequests_NotOwner(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.GetAccessRequestsRequest{ChatId: "test_chat_id"}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(false, nil)

	resp, err := service.GetAccessRequests(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))

	mockStorage.AssertExpectations(t)
}

func TestGrantChatAccess(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.GrantChatAccessRequest{
		ChatId:    "test_chat_id",
		SessionId: "request_session_id",
	}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(true, nil)
	mockStorage.On("GrantChatAccess", mock.Anything, req.ChatId, req.SessionId).Return(nil)

	resp, err := service.GrantChatAccess(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "access_granted", resp.Status)

	mockStorage.AssertExpectations(t)
}

func TestGrantChatAccess_NotOwner(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	sessionID := "test_session_id"
	ctx := createContextWithSession(sessionID)
	req := &pb.GrantChatAccessRequest{
		ChatId:    "test_chat_id",
		SessionId: "request_session_id",
	}

	mockStorage.On("IsChatOwner", mock.Anything, req.ChatId, sessionID).Return(false, nil)

	resp, err := service.GrantChatAccess(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))

	mockStorage.AssertExpectations(t)
}

func TestCreateSession_Error(t *testing.T) {
	mockStorage := new(MockStorage)
	mockProducer := new(MockProducer)
	service := NewChatServiceWithProducer(mockStorage, mockProducer)

	ctx := context.Background()
	req := &pb.CreateSessionRequest{Nickname: "testuser"}

	mockError := status.Error(codes.Internal, "database error")
	mockStorage.On("CreateSession", mock.Anything, mock.AnythingOfType("*models.Session")).Return(mockError)

	resp, err := service.CreateSession(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "database error")

	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Internal, statusErr.Code())

	mockStorage.AssertExpectations(t)
}
