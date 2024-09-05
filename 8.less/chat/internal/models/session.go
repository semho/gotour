package models

import "github.com/google/uuid"

type Session struct {
	ID       string
	Nickname string
}

func NewSession(nickname string) *Session {
	return &Session{
		ID:       uuid.New().String(),
		Nickname: nickname,
	}
}
