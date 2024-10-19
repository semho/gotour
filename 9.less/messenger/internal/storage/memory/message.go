package memory

import (
	"errors"
	"github.com/google/uuid"
	"messenger/internal/model"
	"time"
)

func (db *DB) GetMessageCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.messages)
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

	// существуют ли отправитель
	chat, exists := db.chats[message.ChatID]
	if !exists {
		return nil, errors.New("chat not found")
	}
	// и является ли участником
	isParticipant := false
	for _, participantID := range chat.Participants {
		if participantID == message.SenderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("sender is not a participant of the chat")
	}
	if chat.Type == model.ChatTypeReadOnly && message.SenderID != chat.CreatorID {
		return nil, errors.New("only the creator can send messages in read-only chats")
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

func (db *DB) GetAllMessages() ([]*model.Message, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	messages := make([]*model.Message, 0, len(db.messages))
	for _, message := range db.messages {
		messages = append(messages, message)
	}

	return messages, nil
}
