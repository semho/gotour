package memory

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"messenger/internal/model"
)

type DB struct {
	mu       sync.RWMutex
	users    map[uuid.UUID]*model.User
	messages map[uuid.UUID]*model.Message
	chats    map[uuid.UUID]*model.Chat
}

func NewDB() *DB {
	return &DB{
		users:    make(map[uuid.UUID]*model.User),
		messages: make(map[uuid.UUID]*model.Message),
		chats:    make(map[uuid.UUID]*model.Chat),
	}
}

func (db *DB) CreateUser(user *model.User) (*model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	if _, exists := db.users[user.ID]; exists {
		return nil, errors.New("user already exists")
	}

	db.users[user.ID] = user
	return user, nil
}

func (db *DB) GetUser(id uuid.UUID) (*model.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (db *DB) SendMessage(message *model.Message) (*model.Message, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if message.ID == uuid.Nil {
		message.ID = uuid.New()
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// существуют ли отправитель и получатель
	if _, exists := db.users[message.SenderID]; !exists {
		return nil, errors.New("sender not found")
	}
	if _, exists := db.users[message.ReceiverID]; !exists {
		return nil, errors.New("receiver not found")
	}

	// существует ли чат
	if _, exists := db.chats[message.ChatID]; !exists {
		return nil, errors.New("chat not found")
	}

	db.messages[message.ID] = message
	return message, nil
}

func (db *DB) GetMessage(id uuid.UUID) (*model.Message, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	message, exists := db.messages[id]
	if !exists {
		return nil, errors.New("message not found")
	}

	return message, nil
}
func (db *DB) UpdateMessageStatus(id uuid.UUID, status model.MessageStatus) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	message, exists := db.messages[id]
	if !exists {
		return errors.New("message not found")
	}

	message.Status = status
	return nil
}

func (db *DB) CreateChat(chat *model.Chat) (*model.Chat, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if chat.ID == uuid.Nil {
		chat.ID = uuid.New()
	}

	if chat.CreatedAt.IsZero() {
		chat.CreatedAt = time.Now()
	}

	// существуют ли все участники чата
	for _, participantID := range chat.Participants {
		if _, exists := db.users[participantID]; !exists {
			return nil, errors.New("participant not found")
		}
	}

	db.chats[chat.ID] = chat
	return chat, nil
}

func (db *DB) GetChat(id uuid.UUID) (*model.Chat, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	chat, exists := db.chats[id]
	if !exists {
		return nil, errors.New("chat not found")
	}

	return chat, nil
}

func (db *DB) AddUserToChat(chatID, userID uuid.UUID) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	chat, exists := db.chats[chatID]
	if !exists {
		return errors.New("chat not found")
	}

	if _, exists := db.users[userID]; !exists {
		return errors.New("user not found")
	}

	for _, participant := range chat.Participants {
		if participant == userID {
			return errors.New("user already in chat")
		}
	}

	chat.Participants = append(chat.Participants, userID)
	return nil
}

func (db *DB) RemoveUserFromChat(chatID, userID uuid.UUID) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	chat, exists := db.chats[chatID]
	if !exists {
		return errors.New("chat not found")
	}

	for i, participant := range chat.Participants {
		if participant == userID {
			chat.Participants = append(chat.Participants[:i], chat.Participants[i+1:]...)
			return nil
		}
	}

	return errors.New("user not in chat")
}

func (db *DB) GetChatMessages(chatID uuid.UUID) ([]*model.Message, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	if _, exists := db.chats[chatID]; !exists {
		return nil, errors.New("chat not found")
	}

	var chatMessages []*model.Message
	for _, message := range db.messages {
		if message.ChatID == chatID {
			chatMessages = append(chatMessages, message)
		}
	}

	return chatMessages, nil
}
