package middleware

import (
	"chat/pkg/logger"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

const (
	SessionIDKey ContextKey = "session_id"
)

func AuthInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	logger.Log.Info("AuthInterceptor called for method", "method", info.FullMethod)

	if info.FullMethod == "/chat_v1.ChatService/CreateSession" {
		logger.Log.Info("Skipping auth for CreateSession")
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Log.Error("Metadata not found in context")
		return nil, status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	sessionIDs := md.Get("session_id")
	if len(sessionIDs) == 0 {
		logger.Log.Error("session_id not found in metadata")
		return nil, status.Errorf(codes.Unauthenticated, "session_id is not provided")
	}

	sessionID := sessionIDs[0]
	logger.Log.Info("Found session ID in AuthInterceptor", "session_id", sessionID)

	newCtx := context.WithValue(ctx, SessionIDKey, sessionID)
	resp, err := handler(newCtx, req)
	if err != nil {
		logger.Log.Error("Handler returned error in AuthInterceptor", "error", err)
	}
	return resp, err
}

// вытаскиваем id сессии из контекста
func GetSessionID(ctx context.Context) (string, bool) {
	sessionID, ok := ctx.Value(SessionIDKey).(string)
	return sessionID, ok
}
