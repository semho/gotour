package memory

import (
	"errors"
	"github.com/google/uuid"
	"messenger/internal/model"
	"time"
)

func (db *DB) GetChatCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.chats)
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
