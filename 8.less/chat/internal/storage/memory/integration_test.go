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
	logger.Log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	os.Exit(m.Run())
}

func TestChatServiceIntegration(t *testing.T) {
	storage := NewMemoryStorage(1000, 1000)
	chatService := service.NewChatService(storage)

	// Создание сессии
	createSessionReq := &pb.CreateSessionRequest{Nickname: "testuser"}
	session, err := chatService.CreateSession(context.Background(), createSessionReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, session.Id)

	ctx := createContextWithSession(session.Id)

	// Создание чата
	createChatReq := &pb.CreateChatRequest{
		HistorySize: 100,
		TtlSeconds:  3600,
		ReadOnly:    false,
		Private:     false,
	}
	chat, err := chatService.CreateChat(ctx, createChatReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, chat.Id)

	// Отправка сообщения
	sendMessageReq := &pb.SendMessageRequest{
		ChatId: chat.Id,
		Text:   "Hello, World!",
	}
	message, err := chatService.SendMessage(ctx, sendMessageReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, message.Id)
	assert.Equal(t, "Hello, World!", message.Text)

	// Получение истории чата
	getChatHistoryReq := &pb.GetChatHistoryRequest{
		ChatId: chat.Id,
	}
	history, err := chatService.GetChatHistory(ctx, getChatHistoryReq)
	assert.NoError(t, err)
	assert.Len(t, history.Messages, 1)
	assert.Equal(t, "Hello, World!", history.Messages[0].Text)

	// Тест на установку TTL для чата
	setChatTTLReq := &pb.SetChatTTLRequest{
		ChatId:     chat.Id,
		TtlSeconds: 1800,
	}
	_, err = chatService.SetChatTTL(ctx, setChatTTLReq)
	assert.NoError(t, err)

	// Тест на удаление чата
	deleteChatReq := &pb.DeleteChatRequest{
		ChatId: chat.Id,
	}
	_, err = chatService.DeleteChat(ctx, deleteChatReq)
	assert.NoError(t, err)

	// Проверка, что чат действительно удален
	_, err = chatService.GetChatHistory(ctx, getChatHistoryReq)
	assert.Error(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, statusErr.Code())
}

func TestPrivateChatIntegration(t *testing.T) {
	storage := NewMemoryStorage(1000, 1000)
	chatService := service.NewChatService(storage)

	// Создание двух сессий
	createSessionReq1 := &pb.CreateSessionRequest{Nickname: "user1"}
	session1, _ := chatService.CreateSession(context.Background(), createSessionReq1)
	ctx1 := createContextWithSession(session1.Id)

	createSessionReq2 := &pb.CreateSessionRequest{Nickname: "user2"}
	session2, _ := chatService.CreateSession(context.Background(), createSessionReq2)
	ctx2 := createContextWithSession(session2.Id)

	// Создание приватного чата
	createChatReq := &pb.CreateChatRequest{
		HistorySize: 100,
		TtlSeconds:  3600,
		ReadOnly:    false,
		Private:     true,
	}
	chat, err := chatService.CreateChat(ctx1, createChatReq)
	assert.NoError(t, err)

	// Попытка отправить сообщение от user2 (не должно получиться)
	sendMessageReq := &pb.SendMessageRequest{
		ChatId: chat.Id,
		Text:   "Hello from user2",
	}
	_, err = chatService.SendMessage(ctx2, sendMessageReq)
	assert.Error(t, err)
	statusErr, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.PermissionDenied, statusErr.Code())

	// Запрос доступа для user2
	requestAccessReq := &pb.RequestChatAccessRequest{
		ChatId: chat.Id,
	}
	_, err = chatService.RequestChatAccess(ctx2, requestAccessReq)
	assert.NoError(t, err)

	// Предоставление доступа user2
	grantAccessReq := &pb.GrantChatAccessRequest{
		ChatId:    chat.Id,
		SessionId: session2.Id,
	}
	_, err = chatService.GrantChatAccess(ctx1, grantAccessReq)
	assert.NoError(t, err)

	// Теперь user2 может отправить сообщение
	_, err = chatService.SendMessage(ctx2, sendMessageReq)
	assert.NoError(t, err)
}

func createContextWithSession(sessionID string) context.Context {
	md := metadata.New(map[string]string{"session_id": sessionID})
	ctx := metadata.NewIncomingContext(context.Background(), md)
	ctx, _ = applyAuthInterceptor(ctx)
	return ctx
}

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
