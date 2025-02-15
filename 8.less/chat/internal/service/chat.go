package service

import (
	"chat/internal/middleware"
	"chat/pkg/customerrors"
	"chat/pkg/kafka/producer"
	kafka_v1 "chat/pkg/kafka/v1"
	"chat/pkg/logger"
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

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
	storage  storage.Storage
	producer producer.Producer
}

func NewChatService(storage storage.Storage, kafkaBrokers []string, topic string) (*ChatService, error) {
	prod, err := producer.NewKafkaProducer(kafkaBrokers, topic)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedCreateProducer, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedCreateProducer, err),
		)
	}

	return &ChatService{storage: storage, producer: prod}, nil
}

func NewChatServiceWithProducer(storage storage.Storage, prod producer.Producer) *ChatService {
	return &ChatService{
		storage:  storage,
		producer: prod,
	}
}

func (s *ChatService) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.Session, error) {
	logger.Log.Info("Creating session", "nickname", req.Nickname)
	session := models.NewSession(req.Nickname)
	err := s.storage.CreateSession(ctx, session)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToCreateSession, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToCreateSession, err),
		)
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
		return nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	_, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToGetSession, "error", err)
		return nil, status.Errorf(
			codes.Unauthenticated,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgInvalidSession, err),
		)
	}

	var ttl *time.Time
	if req.TtlSeconds > 0 {
		t := time.Now().Add(time.Duration(req.TtlSeconds) * time.Second)
		ttl = &t
	}

	historySize := int(req.HistorySize)
	if historySize <= 0 {
		historySize = s.storage.GetDefaultHistorySize()
	}

	if historySize > math.MaxInt32 {
		return nil, status.Errorf(codes.InvalidArgument, "history size too large, maximum value is %d", math.MaxInt32)
	}

	chat := models.NewChat(int(req.HistorySize), ttl, req.ReadOnly, req.Private, sessionID)
	err = s.storage.CreateChat(ctx, chat)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToCreateChat, "error", err)
		//nolint:govet
		return nil, status.Errorf(codes.Internal, customerrors.FormatError(customerrors.ErrMsgFailedToCreateChat, err))
	}
	return &pb.Chat{
		Id:          chat.ID,
		HistorySize: int32(historySize),
		Ttl:         timestamppb.New(*chat.TTL),
		ReadOnly:    chat.ReadOnly,
		Private:     chat.Private,
		OwnerId:     chat.OwnerID,
	}, nil
}

func (s *ChatService) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*emptypb.Empty, error) {
	logger.Log.Info("Deleting chat", "ChatId", req.ChatId)
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	isOwner, err := s.storage.IsChatOwner(ctx, req.ChatId, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToCheckChatOwnership, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToCheckChatOwnership, err),
		)
	}
	if !isOwner {
		return nil, status.Errorf(codes.PermissionDenied, customerrors.ErrMsgOnlyChatOwnerCanDeleteChat)
	}

	err = s.storage.DeleteChat(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToDeleteChat, "error", err)
		//nolint:govet
		return nil, status.Errorf(codes.Internal, customerrors.FormatError(customerrors.ErrMsgFailedToDeleteChat, err))
	}
	return &emptypb.Empty{}, nil
}

func (s *ChatService) SetChatTTL(ctx context.Context, req *pb.SetChatTTLRequest) (*emptypb.Empty, error) {
	logger.Log.Info("Set ttl", "TtlSeconds", req.TtlSeconds)
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	isOwner, err := s.storage.IsChatOwner(ctx, req.ChatId, sessionID)
	if err != nil {
		var chatErr *customerrors.ChatError
		if errors.As(err, &chatErr) && errors.Is(chatErr.Err, customerrors.ErrChatNotFound) {
			logger.Log.Error(customerrors.ErrMsgChatNotFoundService, "ChatId", req.ChatId)
			return nil, status.Errorf(
				codes.NotFound,
				//nolint:govet
				customerrors.FormatError(customerrors.ErrMsgChatNotFoundService, err),
			)
		}
		logger.Log.Error(customerrors.ErrMsgFailedToCheckChatOwnership, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToCheckChatOwnership, err),
		)
	}
	if !isOwner {
		return nil, status.Errorf(codes.PermissionDenied, customerrors.ErrMsgOnlyChatOwnerCanSetTTL)
	}

	newTTL := time.Now().Add(time.Duration(req.TtlSeconds) * time.Second)
	err = s.storage.SetChatTTL(ctx, req.ChatId, newTTL)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToSetChatTTL, "error", err)
		//nolint:govet
		return nil, status.Errorf(codes.Internal, customerrors.FormatError(customerrors.ErrMsgFailedToSetChatTTL, err))
	}
	logger.Log.Info("Chat TTL set successfully", "ChatId", req.ChatId, "NewTTL", newTTL)
	return &emptypb.Empty{}, nil
}

func (s *ChatService) validateChatAccess(ctx context.Context, chatID string) (string, *models.Chat, error) {
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		logger.Log.Error(customerrors.ErrMsgSessionNotFound)
		return "", nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	_, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToGetSession, "error", err)
		return "", nil, status.Errorf(
			codes.Unauthenticated,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgInvalidSession, err),
		)
	}

	chat, err := s.storage.GetChat(ctx, chatID)
	if err != nil {
		var chatErr *customerrors.ChatError
		if errors.As(err, &chatErr) {
			if errors.Is(chatErr.Err, customerrors.ErrChatNotFound) {
				logger.Log.Error(customerrors.ErrMsgChatNotFoundService, "chatID", chatID)
				return "", nil, status.Errorf(
					codes.NotFound,
					//nolint:govet
					customerrors.FormatError(customerrors.ErrMsgChatNotFoundService, err),
				)
			}
		}
		logger.Log.Error(customerrors.ErrMsgChatNotFoundService, "error", err)
		return "", nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgChatNotFoundService, err),
		)
	}

	if chat.TTL != nil && time.Now().After(*chat.TTL) {
		return "", nil, status.Errorf(codes.FailedPrecondition, customerrors.ErrMsgChatExpired)
	}

	if chat.Private {
		hasAccess, err := s.storage.HasChatAccess(ctx, chatID, sessionID)
		if err != nil {
			logger.Log.Error(customerrors.ErrMsgFailedToCheckChatAccess, "error", err)
			return "", nil, status.Errorf(
				codes.Internal,
				//nolint:govet
				customerrors.FormatError(customerrors.ErrMsgFailedToCheckChatAccess, err),
			)
		}
		if !hasAccess {
			return "", nil, status.Errorf(codes.PermissionDenied, customerrors.ErrMsgNoAccessToPrivateChat)
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
		return nil, status.Errorf(codes.PermissionDenied, customerrors.ErrMsgChatIsReadOnly)
	}

	session, err := s.storage.GetSession(ctx, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToGetSession, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToGetSession, err),
		)
	}

	// Получаем или генерируем никнейм
	var nickname string
	if session.Nickname == "" {
		if existingNickname, exists := session.AnonNicknames[req.ChatId]; exists {
			nickname = existingNickname
		} else {
			anonCount, err := s.storage.GetAndIncrementAnonCount(ctx, req.ChatId)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to generate anonymous nickname: %v", err)
			}
			nickname = fmt.Sprintf("Аноним #%d", anonCount)

			if err := s.storage.SaveAnonNickname(ctx, req.ChatId, sessionID, nickname); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to save anonymous nickname: %v", err)
			}
		}
	} else {
		nickname = session.Nickname
	}

	message := models.NewMessage(req.ChatId, sessionID, nickname, req.Text)
	// евент под кафка
	event := &kafka_v1.ChatMessageEvent{
		Metadata: &kafka_v1.ChatMessageEvent_Metadata{
			EventId:   uuid.New().String(),
			CreatedAt: timestamppb.Now(),
			EventType: kafka_v1.ChatMessageEvent_EVENT_TYPE_CREATED,
		},
		Payload: &kafka_v1.ChatMessageEvent_Payload{
			MessageId: message.ID,
			ChatId:    message.ChatID,
			SessionId: message.SessionID,
			Nickname:  message.Nickname,
			Text:      message.Text,
			Timestamp: timestamppb.New(message.Timestamp),
		},
	}
	//отправка в кафка
	if err = s.producer.SendMessage(ctx, event); err != nil {
		logger.Log.Error("Send message to kafka", "Error", err)
		return nil, status.Errorf(codes.Internal, "failed to send message to kafka: %v", err)
	}

	return &pb.Message{
		Id:        message.ID,
		ChatId:    message.ChatID,
		SessionId: message.SessionID,
		Nickname:  message.Nickname,
		Text:      message.Text,
		Timestamp: timestamppb.New(message.Timestamp),
	}, nil
}

func (s *ChatService) Close() error {
	if s.producer != nil {
		return s.producer.Close()
	}
	return nil
}

func (s *ChatService) GetChatHistory(ctx context.Context, req *pb.GetChatHistoryRequest) (*pb.ChatHistory, error) {
	logger.Log.Info("Chat history", "ChatId", req.ChatId)

	_, _, err := s.validateChatAccess(ctx, req.ChatId)
	if err != nil {
		return nil, err
	}

	messages, err := s.storage.GetChatHistory(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToGetChatHistory, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToGetChatHistory, err),
		)
	}

	pbMessages := make([]*pb.Message, len(messages))
	for i, msg := range messages {
		pbMessages[i] = &pb.Message{
			Id:        msg.ID,
			ChatId:    msg.ChatID,
			SessionId: msg.SessionID,
			Nickname:  msg.Nickname,
			Text:      msg.Text,
			Timestamp: timestamppb.New(msg.Timestamp),
		}
	}
	return &pb.ChatHistory{Messages: pbMessages}, nil
}

func (s *ChatService) RequestChatAccess(
	ctx context.Context,
	req *pb.RequestChatAccessRequest,
) (*pb.RequestChatAccessResponse, error) {
	logger.Log.Info("Request chat access", "ChatId", req.ChatId)
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	chat, err := s.storage.GetChat(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgChatNotFoundService, "error", err)
		//nolint:govet
		return nil, status.Errorf(codes.NotFound, customerrors.FormatError(customerrors.ErrMsgChatNotFoundService, err))
	}

	if !chat.Private {
		return nil, status.Errorf(codes.FailedPrecondition, customerrors.ErrMsgChatIsNotPrivate)
	}

	hasAccess, err := s.storage.HasChatAccess(ctx, req.ChatId, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToCheckChatAccess, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToCheckChatAccess, err),
		)
	}
	if hasAccess {
		return &pb.RequestChatAccessResponse{Status: "already_has_access"}, nil
	}

	err = s.storage.RequestChatAccess(ctx, req.ChatId, sessionID)
	if err != nil {
		if err != nil {
			switch err {
			case customerrors.ErrAccessAlreadyRequested:
				return &pb.RequestChatAccessResponse{Status: "request_already_sent"}, nil
			case customerrors.ErrAccessAlreadyExist:
				return &pb.RequestChatAccessResponse{Status: "already_has_access"}, nil
			default:
				logger.Log.Error(customerrors.ErrMsgFailedToRequestChatAccess, "error", err)
				return nil, status.Errorf(
					codes.Internal,
					//nolint:govet
					customerrors.FormatError(customerrors.ErrMsgFailedToRequestChatAccess, err),
				)
			}
		}
	}
	return &pb.RequestChatAccessResponse{Status: "request_sent"}, nil
}

func (s *ChatService) GetAccessRequests(ctx context.Context, req *pb.GetAccessRequestsRequest) (
	*pb.AccessRequestList,
	error,
) {
	logger.Log.Info("Get access request", "ChatId", req.ChatId)
	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	isOwner, err := s.storage.IsChatOwner(ctx, req.ChatId, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToCheckChatOwnership, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToCheckChatOwnership, err),
		)
	}
	if !isOwner {
		return nil, status.Errorf(codes.PermissionDenied, customerrors.ErrMsgOnlyChatOwnerCanViewAccessRequests)
	}

	sessionIDs, err := s.storage.GetAccessRequests(ctx, req.ChatId)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToGetAccessRequests, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToGetAccessRequests, err),
		)
	}
	requests := make([]*pb.Session, 0, len(sessionIDs))
	for _, id := range sessionIDs {
		session, err := s.storage.GetSession(ctx, id)
		if err != nil {
			logger.Log.Error(customerrors.ErrMsgFailedToGetSession, "error", err)
			continue
		}
		requests = append(
			requests, &pb.Session{
				Id:       session.ID,
				Nickname: session.Nickname,
			},
		)
	}
	return &pb.AccessRequestList{Requests: requests}, nil
}

func (s *ChatService) GrantChatAccess(ctx context.Context, req *pb.GrantChatAccessRequest) (
	*pb.GrantChatAccessResponse,
	error,
) {
	logger.Log.Info(
		"Grant chat access request",
		"ChatId", req.ChatId,
		"SessionToGrant", req.SessionId,
	)

	sessionID, ok := middleware.GetSessionID(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, customerrors.ErrMsgSessionNotFound)
	}

	if sessionID == req.SessionId {
		return &pb.GrantChatAccessResponse{Status: "already_has_access"}, nil
	}

	isOwner, err := s.storage.IsChatOwner(ctx, req.ChatId, sessionID)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToCheckChatOwnership, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToCheckChatOwnership, err),
		)
	}
	if !isOwner {
		return nil, status.Errorf(codes.PermissionDenied, customerrors.ErrMsgOnlyChatOwnerCanGrantAccess)
	}

	err = s.storage.GrantChatAccess(ctx, req.ChatId, req.SessionId)
	if err != nil {
		logger.Log.Error(customerrors.ErrMsgFailedToGrantChatAccess, "error", err)
		return nil, status.Errorf(
			codes.Internal,
			//nolint:govet
			customerrors.FormatError(customerrors.ErrMsgFailedToGrantChatAccess, err),
		)
	}
	return &pb.GrantChatAccessResponse{Status: "access_granted"}, nil
}
