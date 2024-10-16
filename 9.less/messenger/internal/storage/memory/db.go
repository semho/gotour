package memory

import (
	"github.com/google/uuid"
	"messenger/internal/model"
	"messenger/internal/storage"
	"sync"
)

type DB struct {
	mu          sync.RWMutex
	users       map[uuid.UUID]*model.User
	messages    map[uuid.UUID]*model.Message
	chats       map[uuid.UUID]*model.Chat
	maxUsers    int
	maxMessages int
	maxChats    int
}

func NewDB(maxUsers, maxMessages, maxChats int) *DB {
	return &DB{
		users:       make(map[uuid.UUID]*model.User),
		messages:    make(map[uuid.UUID]*model.Message),
		chats:       make(map[uuid.UUID]*model.Chat),
		maxUsers:    maxUsers,
		maxMessages: maxMessages,
		maxChats:    maxChats,
	}
}

var _ storage.Storage = (*DB)(nil)
