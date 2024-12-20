package memory

import (
	"chat/pkg/customerrors"
	"context"
	"sync"
	"time"

	"chat/internal/models"
)

type Storage struct {
	sessions       map[string]*models.Session
	chats          map[string]*models.Chat
	accessRequests map[string][]string
	anonCounts     map[string]int
	mu             sync.RWMutex
	maxChatSize    int
	maxChatsCount  int
}

func NewMemoryStorage(maxChatSize, maxChatsCount int) *Storage {
	return &Storage{
		sessions:       make(map[string]*models.Session),
		chats:          make(map[string]*models.Chat),
		accessRequests: make(map[string][]string),
		anonCounts:     make(map[string]int),
		maxChatSize:    maxChatSize,
		maxChatsCount:  maxChatsCount,
	}
}

func (s *Storage) GetDefaultHistorySize() int {
	return s.maxChatSize
}

func (s *Storage) GetAndIncrementAnonCount(_ context.Context, chatID string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count, exists := s.anonCounts[chatID]
	if !exists {
		count = 0
	}
	count++
	s.anonCounts[chatID] = count
	return count, nil
}

func (s *Storage) CreateSession(_ context.Context, session *models.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.ID] = session
	return nil
}

func (s *Storage) GetSession(_ context.Context, sessionID string) (*models.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, customerrors.NewSessionError(sessionID, customerrors.ErrSessionNotFound)
	}
	return session, nil
}

func (s *Storage) CreateChat(_ context.Context, chat *models.Chat) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.chats) >= s.maxChatsCount {
		return customerrors.NewChatError(chat.ID, customerrors.ErrMaxNumberReached)
	}

	s.chats[chat.ID] = chat
	return nil
}

func (s *Storage) GetChat(_ context.Context, chatID string) (*models.Chat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return nil, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}
	return chat, nil
}

func (s *Storage) DeleteChat(_ context.Context, chatID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.chats, chatID)
	delete(s.accessRequests, chatID)
	return nil
}

func (s *Storage) SetChatTTL(_ context.Context, chatID string, ttl time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}
	chat.TTL = &ttl
	return nil
}

func (s *Storage) AddMessage(_ context.Context, message *models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, ok := s.chats[message.ChatID]
	if !ok {
		return customerrors.NewChatError(chat.ID, customerrors.ErrChatNotFound)
	}

	if chat.HistorySize > 0 && len(chat.Messages) >= chat.HistorySize {
		if len(chat.Messages) > 0 {
			chat.Messages = chat.Messages[1:]
		} else {
			chat.Messages = []models.Message{}
		}
	}
	chat.Messages = append(chat.Messages, *message)
	return nil
}

func (s *Storage) GetChatHistory(_ context.Context, chatID string) ([]*models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return nil, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	messages := make([]*models.Message, len(chat.Messages))
	for i := range chat.Messages {
		messages[i] = &chat.Messages[i]
	}
	return messages, nil
}

func (s *Storage) RequestChatAccess(_ context.Context, chatID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	if chat.OwnerID == sessionID {
		return customerrors.NewChatError(chatID, customerrors.ErrAccessAlreadyExist)
	}

	for _, id := range chat.AllowedUsers {
		if id == sessionID {
			return customerrors.NewChatError(chatID, customerrors.ErrAccessAlreadyExist)
		}
	}

	for _, id := range s.accessRequests[chatID] {
		if id == sessionID {
			return customerrors.NewChatError(chatID, customerrors.ErrAccessAlreadyRequested)
		}
	}

	s.accessRequests[chatID] = append(s.accessRequests[chatID], sessionID)
	return nil
}

func (s *Storage) GetAccessRequests(_ context.Context, chatID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	requests, ok := s.accessRequests[chatID]
	if !ok {
		return nil, customerrors.NewChatError(chatID, customerrors.ErrChatNotFoundRequests)
	}
	return requests, nil
}

func (s *Storage) GrantChatAccess(_ context.Context, chatID, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}
	chat.AllowedUsers = append(chat.AllowedUsers, sessionID)
	requests := s.accessRequests[chatID]
	for i, id := range requests {
		if id == sessionID {
			s.accessRequests[chatID] = append(requests[:i], requests[i+1:]...)
			break
		}
	}

	return nil
}

func (s *Storage) HasChatAccess(_ context.Context, chatID, sessionID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return false, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	if !chat.Private {
		return true, nil
	}

	if chat.OwnerID == sessionID {
		return true, nil
	}

	for _, id := range chat.AllowedUsers {
		if id == sessionID {
			return true, nil
		}
	}

	return false, nil
}

func (s *Storage) IsChatOwner(_ context.Context, chatID, sessionID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, ok := s.chats[chatID]
	if !ok {
		return false, customerrors.NewChatError(chatID, customerrors.ErrChatNotFound)
	}

	return chat.OwnerID == sessionID, nil
}

func (s *Storage) SaveAnonNickname(_ context.Context, chatID, sessionID, nickname string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return customerrors.NewSessionError(sessionID, customerrors.ErrSessionNotFound)
	}

	if session.AnonNicknames == nil {
		session.AnonNicknames = make(map[string]string)
	}
	session.AnonNicknames[chatID] = nickname
	return nil
}
