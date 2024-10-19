package model

import (
	"github.com/google/uuid"
	"time"
)

type MessageStatus string

const (
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
)

type Message struct {
	ID        uuid.UUID     `json:"id"`
	SenderID  uuid.UUID     `json:"sender_id"`
	ChatID    uuid.UUID     `json:"chat_id"`
	Text      string        `json:"text"`
	Timestamp time.Time     `json:"timestamp"`
	Status    MessageStatus `json:"status"`
}
