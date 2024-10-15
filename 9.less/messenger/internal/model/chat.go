package model

import (
	"github.com/google/uuid"
	"time"
)

type ChatType string

const (
	ChatTypePublic   ChatType = "public"
	ChatTypeReadOnly ChatType = "read_only"
	ChatTypePrivate  ChatType = "private"
)

type Chat struct {
	ID           uuid.UUID   `json:"id"`
	Type         ChatType    `json:"type"`
	Participants []uuid.UUID `json:"participants"`
	CreatedAt    time.Time   `json:"created_at"`
}
