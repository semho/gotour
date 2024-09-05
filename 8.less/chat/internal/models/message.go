package models

import (
	"github.com/google/uuid"
	"time"
)

type Message struct {
	ID        string
	ChatID    string
	SessionID string
	Text      string
	Timestamp time.Time
}

func NewMessage(chatID, sessionID, text string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		ChatID:    chatID,
		SessionID: sessionID,
		Text:      text,
		Timestamp: time.Now(),
	}
}
