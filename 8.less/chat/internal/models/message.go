package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string
	ChatID    string
	SessionID string
	Nickname  string
	Text      string
	Timestamp time.Time
}

func NewMessage(chatID, sessionID, nickname, text string) *Message {
	return &Message{
		ID:        uuid.New().String(),
		ChatID:    chatID,
		SessionID: sessionID,
		Nickname:  nickname,
		Text:      text,
		Timestamp: time.Now(),
	}
}
