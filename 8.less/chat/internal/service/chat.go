package service

import (
	"chat/internal/middleware"
	"chat/pkg/logger"
	"context"
	"time"

	"chat/internal/models"
	"chat/internal/storage"
	pb "chat/pkg/chat/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ChatService struct {
	pb.UnimplementedChatServiceServer
	storage storage.Storage
}

func NewChatService(storage storage.Storage) *ChatService {
	return &ChatService{storage: storage}
}

func (s *ChatService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	logger.Log.Info("Creating session", "nickname", req.Nickname)
	session := models.NewSession(req.Nickname)
	err := s.storage.CreateSession(ctx, session)
	if err != nil {
		logger.Log.Error("Failed to create session", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}
	logger.Log.Info("Session created", "session_id", session.ID)
	return &pb.Session{
		Id:       session.ID,
		Nickname: session.Nickname,
	}, nil
}

func (s *ChatService) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.Chat, error) {
	logger.Log.Info(
		"Creating chat",
		"history_size",
		req.HistorySize,
		"ttl_seconds",
		req.TtlSeconds,
		"read_only",
		req.ReadOnly,
		"private",
		req.Private,
	)
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "session ID not found in context")
	}

	var ttl *time.Time
	if req.TtlSeconds > 0 {
		t := time.Now().Add(time.Duration(req.TtlSeconds) * time.Second)
		ttl = &t
	}
	chat := models.NewChat(int(req.HistorySize), ttl, req.ReadOnly, req.Private, sessionID)
	err := s.storage.CreateChat(ctx, chat)
	if err != nil {
		logger.Log.Error("Failed to create chat", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create chat: %v", err)
	}
	return &pb.Chat{
		Id:          chat.ID,
		HistorySize: req.HistorySize,
		Ttl:         timestamppb.New(*chat.TTL),
		ReadOnly:    chat.ReadOnly,
		Private:     chat.Private,
		OwnerId:     chat.OwnerID,
	}, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*emptypb.Empty, error) {
	logger.Log.Info("Deleting chat", "ChatId", req.ChatId)
	err := s.storage.DeleteChat(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error("Failed to delete chat", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete chat: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) SetChatTTL(ctx context.Context, req *pb.SetChatTTLRequest) (*emptypb.Empty, error) {
	logger.Log.Info("Set ttl", "TtlSeconds", req.TtlSeconds)
	ttl := time.Now().Add(time.Duration(req.TtlSeconds) * time.Second)
	err := s.storage.SetChatTTL(ctx, req.ChatId, ttl)
	if err != nil {
		logger.Log.Error("failed to set chat TTL:", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to set chat TTL: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) validateChatAccess(ctx context.Context, chatID string) (string, *models.Chat, error) {
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		logger.Log.Error("Session ID not found in context")
		return "", nil, status.Errorf(codes.Unauthenticated, "session ID not found in context")
	}

	_, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		logger.Log.Error("Failed to get session", "error", err)
		return "", nil, status.Errorf(codes.Unauthenticated, "invalid session: %v", err)
	}

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		logger.Log.Error("Failed to get chat", "error", err)
		return "", nil, status.Errorf(codes.NotFound, "chat not found: %v", err)
	}

	if chat.TTL != nil && time.Now().After(*chat.TTL) {
		return "", nil, status.Errorf(codes.FailedPrecondition, "chat has expired")
	}

	if chat.Private {
		hasAccess, err := s.storage.HasChatAccess(ctx, chatID, sessionID)
		if err != nil {
			logger.Log.Error("Failed to check chat access", "error", err)
			return "", nil, status.Errorf(codes.Internal, "failed to check chat access: %v", err)
		}
		if !hasAccess {
			return "", nil, status.Errorf(codes.PermissionDenied, "no access to private chat")
		}
	}

	return sessionID, chat, nil
}

func (s *ChatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	logger.Log.Info("Send message", "Text", req.Text)
	sessionID, chat, err := s.validateChatAccess(ctx, req.ChatId)
	if err != nil {
		return nil, err
	}

	if chat.ReadOnly && chat.OwnerID != sessionID {
		return nil, status.Errorf(codes.PermissionDenied, "chat is read-only")
	}

	message := models.NewMessage(req.ChatId, sessionID, req.Text)
	err = s.storage.AddMessage(ctx, message)
	if err != nil {
		logger.Log.Error("failed to send messag:", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}
	return &pb.Message{
		Id:        message.ID,
		ChatId:    message.ChatID,
		SessionId: message.SessionID,
		Text:      message.Text,
		Timestamp: timestamppb.New(message.Timestamp),
	}, nil
}

func (s *ChatService) GetChatHistory(ctx context.Context, req *pb.GetChatHistoryRequest) (*pb.ChatHistory, error) {
	logger.Log.Info("Chat history", "ChatId", req.ChatId)

	_, _, err := s.validateChatAccess(ctx, req.ChatId)
	if err != nil {
		return nil, err
	}

	messages, err := s.storage.GetChatHistory(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error("failed to get chat history:", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get chat history: %v", err)
	}

	pbMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.Message{
			Id:        msg.ID,
			ChatId:    msg.ChatID,
			SessionId: msg.SessionID,
			Text:      msg.Text,
			Timestamp: timestamppb.New(msg.Timestamp),
		}
	}
	return &pb.ChatHistory{Messages: pbMessages}, nil
}

func (s *ChatService) RequestChatAccess(ctx context.Context, req *pb.RequestChatAccessRequest) (*emptypb.Empty, error) {
	logger.Log.Info("Request chat access", "ChatId", req.ChatId)
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "session ID not found in context")
	}

	err := s.storage.RequestChatAccess(ctx, req.ChatId, sessionID)
	if err != nil {
		logger.Log.Error("failed to request chat access:", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to request chat access: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) GetAccessRequests(ctx context.Context, req *pb.GetAccessRequestsRequest) (
	*pb.AccessRequestList,
	error,
) {
	logger.Log.Info("Get access request", "ChatId", req.ChatId)
	sessionIDs, err := s.storage.GetAccessRequests(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error("failed to get access requests:", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to get access requests: %v", err)
	}
	requests := make([]*pb.Session, len(sessionIDs))
	for i, id := range sessionIDs {
		session, err := s.storage.GetSession(ctx, id)
		if err != nil {
			logger.Log.Error("failed to get session info:", "error", err)
			return nil, status.Errorf(codes.Internal, "failed to get session info: %v", err)
		}
		requests[i] = &pb.Session{
			Id:       session.ID,
			Nickname: session.Nickname,
		}
	}
	return &pb.AccessRequestList{Requests: requests}, nil
}

func (s *ChatService) GrantChatAccess(ctx context.Context, req *pb.GrantChatAccessRequest) (*emptypb.Empty, error) {
	logger.Log.Info("Grant chat access", "ChatId", req.ChatId)
	err := s.storage.GrantChatAccess(ctx, req.ChatId, req.SessionId)
	if err != nil {
		logger.Log.Error("failed to grant chat access:", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to grant chat access: %v", err)
	}
	return &emptypb.Empty{}, nil
}
