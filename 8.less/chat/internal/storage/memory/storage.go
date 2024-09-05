//package memory
//
//import (
//	"context"
//	"sync"
//	"time"
//
//	"chat/internal/models"
//	"chat/internal/storage"
//)
//
//type MemoryStorage struct {
//	sessions       map[string]*models.Session
//	chats          map[string]*models.Chat
//	accessRequests map[string][]string
//	mu             sync.RWMutex
//}
//
//func NewMemoryStorage() *MemoryStorage {
//	return &MemoryStorage{
//		sessions:       make(map[string]*models.Session),
//		chats:          make(map[string]*models.Chat),
//		accessRequests: make(map[string][]string),
//	}
//}
//
//func (s *MemoryStorage) CreateSession(ctx context.Context, session *models.Session) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	s.sessions[session.ID] = session
//	return nil
//}
//
//func (s *MemoryStorage) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	session, ok := s.sessions[sessionID]
//	if !ok {
//		return nil, storage.ErrSessionNotFound
//	}
//	return session, nil
//}
//
//func (s *MemoryStorage) CreateChat(ctx context.Context, chat *models.Chat) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	s.chats[chat.ID] = chat
//	return nil
//}
//
//func (s *MemoryStorage) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	chat, ok := s.chats[chatID]
//	if !ok {
//		return nil, storage.ErrChatNotFound
//	}
//	return chat, nil
//}
//
//func (s *MemoryStorage) DeleteChat(ctx context.Context, chatID string) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	delete(s.chats, chatID)
//	return nil
//}
//
//func (s *MemoryStorage) SetChatTTL(ctx context.Context, chatID string, ttl time.Time) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	chat, ok := s.chats[chatID]
//	if !ok {
//		return storage.ErrChatNotFound
//	}
//	chat.TTL = &ttl
//	return nil
//}
//
//func (s *MemoryStorage) AddMessage(ctx context.Context, message *models.Message) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	chat, ok := s.chats[message.ChatID]
//	if !ok {
//		return storage.ErrChatNotFound
//	}
//	chat.Messages = append(chat.Messages, *message)
//	if len(chat.Messages) > chat.HistorySize {
//		chat.Messages = chat.Messages[len(chat.Messages)-chat.HistorySize:]
//	}
//	return nil
//}
//
//func (s *MemoryStorage) GetChatHistory(ctx context.Context, chatID string) ([]models.Message, error) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	chat, ok := s.chats[chatID]
//	if !ok {
//		return nil, storage.ErrChatNotFound
//	}
//	return chat.Messages, nil
//}
//
//func (s *MemoryStorage) RequestChatAccess(ctx context.Context, chatID, sessionID string) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	s.accessRequests[chatID] = append(s.accessRequests[chatID], sessionID)
//	return nil
//}
//
//func (s *MemoryStorage) GetAccessRequests(ctx context.Context, chatID string) ([]string, error) {
//	s.mu.RLock()
//	defer s.mu.RUnlock()
//	return s.accessRequests[chatID], nil
//}
//
//func (s *MemoryStorage) GrantChatAccess(ctx context.Context, chatID, sessionID string) error {
//	s.mu.Lock()
//	defer s.mu.Unlock()
//	//chat, ok := s.chats[chatID]
//	_, ok := s.chats[chatID]
//	if !ok {
//		return storage.ErrChatNotFound
//	}
//	//TODO: тут надо реализовать логику доступа, используя chat, сейчас удалили сессию из запросов на доступ
//	requests := s.accessRequests[chatID]
//	for i, id := range requests {
//		if id == sessionID {
//			s.accessRequests[chatID] = append(requests[:i], requests[i+1:]...)
//			break
//		}
//	}
//	return nil
//}

package memory

import (
	"context"
	"sync"
	"time"

	"chat/internal/models"
	"chat/internal/storage"
)

type MemoryStorage struct {
	sessions       map[string]*models.Session
	chats          map[string]*models.Chat
	accessRequests map[string][]string
	mu             sync.RWMutex
	maxChatSize    int
	maxChatsCount  int
}

func NewMemoryStorage(maxChatSize, maxChatsCount int) *MemoryStorage {
	return &MemoryStorage{
		sessions:       make(map[string]*models.Session),
		chats:          make(map[string]*models.Chat),
		accessRequests: make(map[string][]string),
		maxChatSize:    maxChatSize,
		maxChatsCount:  maxChatsCount,
	}
}

func (s *MemoryStorage) CreateSession(ctx context.Context, session *models.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

func (s *MemoryStorage) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, storage.ErrSessionNotFound
	}
	return session, nil
}

func (s *MemoryStorage) CreateChat(ctx context.Context, chat *models.Chat) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.chats) >= s.maxChatsCount {
		return storage.ErrMaxNumberReached
	}

	s.chats[chat.ID] = chat
	return nil
}

func (s *MemoryStorage) GetChat(ctx context.Context, chatID string) (*models.Chat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return nil, storage.ErrChatNotFound
	}
	return chat, nil
}

func (s *MemoryStorage) DeleteChat(ctx context.Context, chatID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.chats, chatID)
	delete(s.accessRequests, chatID)
	return nil
}

func (s *MemoryStorage) SetChatTTL(ctx context.Context, chatID string, ttl time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return storage.ErrChatNotFound
	}
	chat.TTL = &ttl
	return nil
}

func (s *MemoryStorage) AddMessage(ctx context.Context, message *models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, ok := s.chats[message.ChatID]
	if !ok {
		return storage.ErrChatNotFound
	}

	if len(chat.Messages) >= chat.HistorySize {
		chat.Messages = chat.Messages[1:]
	}
	chat.Messages = append(chat.Messages, *message)
	return nil
}

func (s *MemoryStorage) GetChatHistory(ctx context.Context, chatID string) ([]*models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return nil, storage.ErrChatNotFound
	}

	messages := make([]*models.Message, len(chat.Messages))
	for i := range chat.Messages {
		messages[i] = &chat.Messages[i]
	}
	return messages, nil
}

func (s *MemoryStorage) RequestChatAccess(ctx context.Context, chatID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.accessRequests[chatID] = append(s.accessRequests[chatID], sessionID)
	return nil
}

func (s *MemoryStorage) GetAccessRequests(ctx context.Context, chatID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests, ok := s.accessRequests[chatID]
	if !ok {
		return nil, storage.ErrChatNotFound
	}
	return requests, nil
}

func (s *MemoryStorage) GrantChatAccess(ctx context.Context, chatID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	//chat, ok := s.chats[chatID]
	_, ok := s.chats[chatID]
	if !ok {
		return storage.ErrChatNotFound
	}
	//TODO: тут надо реализовать логику доступа, используя chat, сейчас удалили сессию из запросов на доступ
	//TODO: а лучше контроль доступа должен быть реализован на уровне сервиса
	requests := s.accessRequests[chatID]
	for i, id := range requests {
		if id == sessionID {
			s.accessRequests[chatID] = append(requests[:i], requests[i+1:]...)
			break
		}
	}

	return nil
}
