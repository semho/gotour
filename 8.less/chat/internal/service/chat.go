package service

import (
	"chat/internal/middleware"
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
	session := models.NewSession(req.Nickname)
	err := s.storage.CreateSession(ctx, session)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}
	return &pb.Session{
		Id:       session.ID,
		Nickname: session.Nickname,
	}, nil
}

func (s *ChatService) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.Chat, error) {
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
		return nil, status.Errorf(codes.Internal, "failed to create chat: %v", err)
	}
	return &pb.Chat{
		Id:          chat.ID,
		HistorySize: int32(chat.HistorySize),
		Ttl:         timestamppb.New(*chat.TTL),
		ReadOnly:    chat.ReadOnly,
		Private:     chat.Private,
		OwnerId:     chat.OwnerID,
	}, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*emptypb.Empty, error) {
	err := s.storage.DeleteChat(ctx, req.ChatId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete chat: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) SetChatTTL(ctx context.Context, req *pb.SetChatTTLRequest) (*emptypb.Empty, error) {
	ttl := time.Now().Add(time.Duration(req.TtlSeconds) * time.Second)
	err := s.storage.SetChatTTL(ctx, req.ChatId, ttl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set chat TTL: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.Message, error) {
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "session ID not found in context")
	}

	message := models.NewMessage(req.ChatId, sessionID, req.Text)
	err := s.storage.AddMessage(ctx, message)
	if err != nil {
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
	messages, err := s.storage.GetChatHistory(ctx, req.ChatId)
	if err != nil {
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
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "session ID not found in context")
	}

	err := s.storage.RequestChatAccess(ctx, req.ChatId, sessionID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to request chat access: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) GetAccessRequests(ctx context.Context, req *pb.GetAccessRequestsRequest) (
	*pb.AccessRequestList,
	error,
) {
	sessionIDs, err := s.storage.GetAccessRequests(ctx, req.ChatId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get access requests: %v", err)
	}
	requests := make([]*pb.Session, len(sessionIDs))
	for i, id := range sessionIDs {
		session, err := s.storage.GetSession(ctx, id)
		if err != nil {
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
	err := s.storage.GrantChatAccess(ctx, req.ChatId, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to grant chat access: %v", err)
	}
	return &emptypb.Empty{}, nil
}