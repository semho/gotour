package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	sessionIDKey contextKey = "session_id"
)

func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// пропуск проверки для метода CreateSession, чтобы можно было создать сессию без самой сессии в контексте
	if info.FullMethod == "/chat_v1.ChatService/CreateSession" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	sessionIDs := md.Get("session_id")
	if len(sessionIDs) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "session_id is not provided")
	}

	sessionID := sessionIDs[0]
	//TODO: можно добавить доп проверку id сессии

	newCtx := context.WithValue(ctx, sessionIDKey, sessionID)
	return handler(newCtx, req)
}

// вытаскиваем id сессии из контекста
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(sessionIDKey).(string)
	return sessionID, ok
}
