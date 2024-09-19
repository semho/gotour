package memory

import (
	"chat/internal/models"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndGetSession(t *testing.T) {
	s := NewMemoryStorage(1000, 1000)
	ctx := context.Background()

	session := &models.Session{
		ID:       "test_id",
		Nickname: "test_user",
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
		ID:          "test_chat_id",
		HistorySize: 100,
		ReadOnly:    false,
		Private:     true,
		OwnerID:     "test_owner_id",
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
		ID:          "test_chat_id",
		HistorySize: 3,
		ReadOnly:    false,
		Private:     false,
		OwnerID:     "test_owner_id",
	}
	err := s.CreateChat(ctx, chat)
	assert.NoError(t, err)

	messages := []*models.Message{
		{ID: "msg1", ChatID: chat.ID, SessionID: "session1", Text: "Hello", Timestamp: time.Now()},
		{ID: "msg2", ChatID: chat.ID, SessionID: "session2", Text: "World", Timestamp: time.Now()},
		{ID: "msg3", ChatID: chat.ID, SessionID: "session3", Text: "!", Timestamp: time.Now()},
		{ID: "msg4", ChatID: chat.ID, SessionID: "session4", Text: "Overflow", Timestamp: time.Now()},
	}

	for _, msg := range messages {
		err = s.AddMessage(ctx, msg)
		assert.NoError(t, err)
	}

	history, err := s.GetChatHistory(ctx, chat.ID)
	assert.NoError(t, err)
	assert.Len(t, history, 3, "История чата должна содержать только 3 последних сообщения")

	// Проверяем, что в истории остались только последние 3 сообщения
	assert.Equal(t, messages[1].ID, history[0].ID, "Первое сообщение в истории должно быть вторым отправленным")
	assert.Equal(t, messages[2].ID, history[1].ID, "Второе сообщение в истории должно быть третьим отправленным")
	assert.Equal(t, messages[3].ID, history[2].ID, "Третье сообщение в истории должно быть четвертым отправленным")
}
