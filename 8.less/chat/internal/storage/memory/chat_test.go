package memory

import (
	"chat/internal/models"
	"chat/pkg/customerrors"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndGetSession(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	session := &models.Session{
		ID:            "test_id",
		Nickname:      "test_user",
		AnonNicknames: make(map[string]string),
	}

	err := s.CreateSession(ctx, session)
	assert.NoError(t, err)

	retrievedSession, err := s.GetSession(ctx, session.ID)
	assert.NoError(t, err)
	assert.Equal(t, session, retrievedSession)
}

func TestCreateAndGetChat(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chat := &models.Chat{
		ID:           "test_chat_id",
		HistorySize:  100,
		ReadOnly:     false,
		Private:      true,
		OwnerID:      "test_owner_id",
		AllowedUsers: []string{},
	}

	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	retrievedChat, err := s.GetChat(ctx, chat.ID)
	assert.NoError(t, err)
	assert.Equal(t, chat, retrievedChat)
}

func TestAddAndGetMessage(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chat := &models.Chat{
		ID:           "test_chat_id",
		HistorySize:  3,
		ReadOnly:     false,
		Private:      false,
		OwnerID:      "test_owner_id",
		AllowedUsers: []string{},
	}
	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	messages := []models.Message{
		{ID: "msg1", ChatID: chat.ID, SessionID: "session1", Text: "Hello", Timestamp: time.Now()},
		{ID: "msg2", ChatID: chat.ID, SessionID: "session2", Text: "World", Timestamp: time.Now()},
		{ID: "msg3", ChatID: chat.ID, SessionID: "session3", Text: "!", Timestamp: time.Now()},
		{ID: "msg4", ChatID: chat.ID, SessionID: "session4", Text: "Overflow", Timestamp: time.Now()},
	}

	for _, msg := range messages {
		err = s.AddMessage(ctx, &msg)
		assert.NoError(t, err)
	}

	history, err := s.GetChatHistory(ctx, chat.ID)
	assert.NoError(t, err)
	assert.Len(t, history, 3, "История чата должна содержать только 3 последних сообщения")

	assert.Equal(t, messages[1].ID, history[0].ID, "Первое сообщение в истории должно быть вторым отправленным")
	assert.Equal(t, messages[2].ID, history[1].ID, "Второе сообщение в истории должно быть третьим отправленным")
	assert.Equal(t, messages[3].ID, history[2].ID, "Третье сообщение в истории должно быть четвертым отправленным")
}

func TestRequestAndGrantChatAccess(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chat := &models.Chat{
		ID:           "test_chat_id",
		HistorySize:  100,
		ReadOnly:     false,
		Private:      true,
		OwnerID:      "owner_id",
		AllowedUsers: []string{},
	}
	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	// Запрос доступа
	err = s.RequestChatAccess(ctx, chat.ID, "user1")
	assert.NoError(t, err)

	requests, err := s.GetAccessRequests(ctx, chat.ID)
	assert.NoError(t, err)
	assert.Contains(t, requests, "user1")

	// Предоставление доступа
	err = s.GrantChatAccess(ctx, chat.ID, "user1")
	assert.NoError(t, err)

	updatedChat, err := s.GetChat(ctx, chat.ID)
	assert.NoError(t, err)
	assert.Contains(t, updatedChat.AllowedUsers, "user1")

	// Проверка доступа
	hasAccess, err := s.HasChatAccess(ctx, chat.ID, "user1")
	assert.NoError(t, err)
	assert.True(t, hasAccess)
}

func TestIsChatOwner(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chat := &models.Chat{
		ID:           "test_chat_id",
		HistorySize:  100,
		ReadOnly:     false,
		Private:      true,
		OwnerID:      "owner_id",
		AllowedUsers: []string{},
	}
	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	isOwner, err := s.IsChatOwner(ctx, chat.ID, "owner_id")
	assert.NoError(t, err)
	assert.True(t, isOwner)

	isOwner, err = s.IsChatOwner(ctx, chat.ID, "not_owner")
	assert.NoError(t, err)
	assert.False(t, isOwner)
}

func TestGetAndIncrementAnonCount(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chatID := "test_chat_id"

	count, err := s.GetAndIncrementAnonCount(ctx, chatID)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)

	count, err = s.GetAndIncrementAnonCount(ctx, chatID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestDeleteChat(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chat := &models.Chat{
		ID:           "test_chat_id",
		HistorySize:  100,
		ReadOnly:     false,
		Private:      true,
		OwnerID:      "owner_id",
		AllowedUsers: []string{},
	}
	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	err = s.DeleteChat(ctx, chat.ID)
	assert.NoError(t, err)

	_, err = s.GetChat(ctx, chat.ID)
	assert.Error(t, err)

	assert.IsType(t, &customerrors.ChatError{}, err, "Ошибка должна быть типа *customerrors.ChatError")

	chatErr, _ := err.(*customerrors.ChatError)
	assert.Equal(t, chat.ID, chatErr.ChatID)
	assert.Equal(t, customerrors.ErrChatNotFound, chatErr.Err)
}

func TestSetChatTTL(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	chat := &models.Chat{
		ID:           "test_chat_id",
		HistorySize:  100,
		ReadOnly:     false,
		Private:      true,
		OwnerID:      "owner_id",
		AllowedUsers: []string{},
	}
	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	ttl := time.Now().Add(1 * time.Hour)
	err = s.SetChatTTL(ctx, chat.ID, ttl)
	assert.NoError(t, err)

	updatedChat, err := s.GetChat(ctx, chat.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedChat.TTL)
	assert.Equal(t, ttl.Unix(), updatedChat.TTL.Unix())
}
