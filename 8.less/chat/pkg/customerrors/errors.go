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

const (
	ErrMsgSessionNotFound                    = "session ID not found in context"
	ErrMsgInvalidSession                     = "invalid session"
	ErrMsgFailedToGetSession                 = "failed to get session info"
	ErrMsgChatNotFoundService                = "Chat not found in service"
	ErrMsgFailedToCreateChat                 = "failed to create chat"
	ErrMsgFailedToDeleteChat                 = "failed to delete chat"
	ErrMsgFailedToSetChatTTL                 = "failed to set chat TTL"
	ErrMsgChatExpired                        = "chat has expired"
	ErrMsgNoAccessToPrivateChat              = "no access to private chat"
	ErrMsgChatIsReadOnly                     = "chat is read-only"
	ErrMsgFailedToSendMessage                = "failed to send message"
	ErrMsgFailedToGetChatHistory             = "failed to get chat history"
	ErrMsgChatIsNotPrivate                   = "chat is not private"
	ErrMsgFailedToCheckChatAccess            = "failed to check chat access"
	ErrMsgFailedToRequestChatAccess          = "failed to request chat access"
	ErrMsgFailedToCheckChatOwnership         = "failed to check chat ownership"
	ErrMsgOnlyChatOwnerCanDeleteChat         = "only chat owner can delete the chat"
	ErrMsgOnlyChatOwnerCanSetTTL             = "only chat owner can set TTL"
	ErrMsgOnlyChatOwnerCanViewAccessRequests = "only chat owner can view access requests"
	ErrMsgOnlyChatOwnerCanGrantAccess        = "only chat owner can grant access"
	ErrMsgFailedToGetAccessRequests          = "failed to get access requests"
	ErrMsgFailedToGrantChatAccess            = "failed to grant chat access"
)

func FormatError(baseMsg string, err error) string {
	return fmt.Sprintf("%s: %v", baseMsg, err)
}

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
