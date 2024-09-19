package memory

import (
	"chat/internal/middleware"
	"chat/internal/service"
	pb "chat/pkg/chat/v1"
	"chat/pkg/logger"
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestMain(m *testing.M) {
	// глобальный логгер перед запуском тестов
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Запускаем тесты
	os.Exit(m.Run())
}
func TestChatServiceIntegration(t *testing.T) {
	storage := NewMemoryStorage(1000, 1000)
	chatService := service.NewChatService(storage)

	createSessionReq := &pb.CreateSessionRequest{Nickname: "testuser"}
	session, err := chatService.CreateSession(context.Background(), createSessionReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, session.Id)

	// контекст с эмуляцией gRPC метаданных
	md := metadata.New(map[string]string{"session_id": session.Id})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// Применяем AuthInterceptor к контексту
	ctx, err = applyAuthInterceptor(ctx)
	assert.NoError(t, err)

	// session ID корректно добавлен в контекст?
	sessionID, ok := middleware.GetSessionID(ctx)
	assert.True(t, ok, "Session ID should be present in the context")
	assert.Equal(t, session.Id, sessionID, "Session ID in context should match the created session ID")

	createChatReq := &pb.CreateChatRequest{
		HistorySize: 100,
		TtlSeconds:  3600,
		ReadOnly:    false,
		Private:     false,
	}
	chat, err := chatService.CreateChat(ctx, createChatReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, chat.Id)

	// Отправляем сообщение
	sendMessageReq := &pb.SendMessageRequest{
		ChatId: chat.Id,
		Text:   "Hello, World!",
	}
	message, err := chatService.SendMessage(ctx, sendMessageReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, message.Id)
	assert.Equal(t, "Hello, World!", message.Text)

	// история чата
	getChatHistoryReq := &pb.GetChatHistoryRequest{
		ChatId: chat.Id,
	}
	history, err := chatService.GetChatHistory(ctx, getChatHistoryReq)
	assert.NoError(t, err)
	assert.Len(t, history.Messages, 1)
	assert.Equal(t, "Hello, World!", history.Messages[0].Text)
}

// applyAuthInterceptor эмулирует работу AuthInterceptor
func applyAuthInterceptor(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	sessionIDs := md.Get("session_id")
	if len(sessionIDs) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "session_id is not provided")
	}

	sessionID := sessionIDs[0]

	return context.WithValue(ctx, middleware.SessionIDKey, sessionID), nil
}
