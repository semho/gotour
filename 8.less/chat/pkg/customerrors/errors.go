package customerrors

import (
	"errors"
	"fmt"
)

var (
	ErrSessionNotFound        = errors.New("session not found")
	ErrChatNotFound           = errors.New("chat not found")
	ErrMaxNumberReached       = errors.New("maximum number of chats reached")
	ErrAccessAlreadyExist     = errors.New("access already exists")
	ErrAccessAlreadyRequested = errors.New("access already requested")
)

type ChatError struct {
	ChatID string
	Err    error
}

func (e *ChatError) Error() string {
	return fmt.Sprintf("chat error (ID: %s): %v", e.ChatID, e.Err)
}

func (e *ChatError) Unwrap() error {
	return e.Err
}

func NewChatError(chatID string, err error) *ChatError {
	return &ChatError{ChatID: chatID, Err: err}
}

type SessionError struct {
	SessionID string
	Err       error
}

func (e *SessionError) Error() string {
	return fmt.Sprintf("session error (ID: %s): %v", e.SessionID, e.Err)
}

func (e *SessionError) Unwrap() error {
	return e.Err
}

func NewSessionError(sessionID string, err error) *SessionError {
	return &SessionError{SessionID: sessionID, Err: err}
}
