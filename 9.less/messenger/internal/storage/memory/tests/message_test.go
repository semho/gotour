package tests

import (
	"messenger/internal/model"
	"messenger/internal/storage/memory"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSendMessage(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	message := &model.Message{
		ID:        uuid.New(),
		SenderID:  user.ID,
		ChatID:    chat.ID,
		Text:      "Test message",
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}

	sentMessage, err := db.SendMessage(message)
	assert.NoError(t, err)
	assert.Equal(t, message, sentMessage)

	fetchedMessage, err := db.GetMessage(message.ID)
	assert.NoError(t, err)
	assert.Equal(t, message, fetchedMessage)
}

func TestGetMessageCount(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	assert.Equal(t, 0, db.GetMessageCount())

	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	message := &model.Message{
		ID:        uuid.New(),
		SenderID:  user.ID,
		ChatID:    chat.ID,
		Text:      "Test message",
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}
	_, err = db.SendMessage(message)
	assert.NoError(t, err)

	assert.Equal(t, 1, db.GetMessageCount())
}

func TestGetMessage(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	message := &model.Message{
		ID:        uuid.New(),
		SenderID:  user.ID,
		ChatID:    chat.ID,
		Text:      "Test message",
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}
	_, err = db.SendMessage(message)
	assert.NoError(t, err)

	fetchedMessage, err := db.GetMessage(message.ID)
	assert.NoError(t, err)
	assert.Equal(t, message, fetchedMessage)

	_, err = db.GetMessage(uuid.New())
	assert.Error(t, err)
}

func TestUpdateMessageStatus(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	message := &model.Message{
		ID:        uuid.New(),
		SenderID:  user.ID,
		ChatID:    chat.ID,
		Text:      "Test message",
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}
	_, err = db.SendMessage(message)
	assert.NoError(t, err)

	err = db.UpdateMessageStatus(message.ID, model.MessageStatusRead)
	assert.NoError(t, err)

	updatedMessage, err := db.GetMessage(message.ID)
	assert.NoError(t, err)
	assert.Equal(t, model.MessageStatusRead, updatedMessage.Status)

	err = db.UpdateMessageStatus(uuid.New(), model.MessageStatusRead)
	assert.Error(t, err)
}

func TestGetAllMessages(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	message1 := &model.Message{
		ID:        uuid.New(),
		SenderID:  user.ID,
		ChatID:    chat.ID,
		Text:      "Test message 1",
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}
	_, err = db.SendMessage(message1)
	assert.NoError(t, err)

	message2 := &model.Message{
		ID:        uuid.New(),
		SenderID:  user.ID,
		ChatID:    chat.ID,
		Text:      "Test message 2",
		Timestamp: time.Now(),
		Status:    model.MessageStatusSent,
	}
	_, err = db.SendMessage(message2)
	assert.NoError(t, err)

	messages, err := db.GetAllMessages()
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Contains(t, messages, message1)
	assert.Contains(t, messages, message2)
}
