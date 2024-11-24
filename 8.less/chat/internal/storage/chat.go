package storage

import (
	"chat/internal/models"
	"context"
	"time"
)

type Storage interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	GetDefaultHistorySize() int

	CreateChat(ctx context.Context, chat *models.Chat) error
	GetChat(ctx context.Context, chatID string) (*models.Chat, error)
	DeleteChat(ctx context.Context, chatID string) error
	SetChatTTL(ctx context.Context, chatID string, ttl time.Time) error

	AddMessage(ctx context.Context, message *models.Message) error
	GetChatHistory(ctx context.Context, chatID string) ([]*models.Message, error)
	GetAndIncrementAnonCount(ctx context.Context, chatID string) (int, error)
	SaveAnonNickname(ctx context.Context, chatID, sessionID, nickname string) error

	RequestChatAccess(ctx context.Context, chatID, sessionID string) error
	GetAccessRequests(ctx context.Context, chatID string) ([]string, error)
	GrantChatAccess(ctx context.Context, chatID, sessionID string) error
	HasChatAccess(ctx context.Context, chatID, sessionID string) (bool, error)
	IsChatOwner(ctx context.Context, chatID, sessionID string) (bool, error)
}
