package storage

import (
	"chat/internal/models"
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound        = errors.New("session not found")
	ErrChatNotFound           = errors.New("chat not found")
	ErrMaxNumberReached       = errors.New("maximum number of chats reached")
	ErrAccessAlreadyExist     = errors.New("access already exists")
	ErrAccessAlreadyRequested = errors.New("access already requested")
)

type Storage interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)

	CreateChat(ctx context.Context, chat *models.Chat) error
	GetChat(ctx context.Context, chatID string) (*models.Chat, error)
	DeleteChat(ctx context.Context, chatID string) error
	SetChatTTL(ctx context.Context, chatID string, ttl time.Time) error

	AddMessage(ctx context.Context, message *models.Message) error
	GetChatHistory(ctx context.Context, chatID string) ([]*models.Message, error)

	RequestChatAccess(ctx context.Context, chatID, sessionID string) error
	GetAccessRequests(ctx context.Context, chatID string) ([]string, error)
	GrantChatAccess(ctx context.Context, chatID, sessionID string) error
	HasChatAccess(ctx context.Context, chatID, sessionID string) (bool, error)
	IsChatOwner(ctx context.Context, chatID, sessionID string) (bool, error)
}
