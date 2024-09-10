package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID          string
	HistorySize int
	TTL         *time.Time
	ReadOnly    bool
	Private     bool
	OwnerID     string
	Messages    []Message
}

func NewChat(historySize int, ttl *time.Time, readOnly, private bool, ownerID string) *Chat {
	return &Chat{
		ID:          uuid.New().String(),
		HistorySize: historySize,
		TTL:         ttl,
		ReadOnly:    readOnly,
		Private:     private,
		OwnerID:     ownerID,
		Messages:    make([]Message, 0),
	}
}
