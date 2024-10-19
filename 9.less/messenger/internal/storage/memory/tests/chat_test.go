package tests

import (
	"messenger/internal/model"
	"messenger/internal/storage/memory"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateChat(t *testing.T) {
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

	createdChat, err := db.CreateChat(chat)
	assert.NoError(t, err)
	assert.Equal(t, chat, createdChat)

	fetchedChat, err := db.GetChat(chat.ID)
	assert.NoError(t, err)
	assert.Equal(t, chat, fetchedChat)
}

func TestGetChatCount(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	assert.Equal(t, 0, db.GetChatCount())

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

	assert.Equal(t, 1, db.GetChatCount())
}

func TestIsUserInChat(t *testing.T) {
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

	isInChat, err := db.IsUserInChat(chat.ID, user.ID)
	assert.NoError(t, err)
	assert.True(t, isInChat)

	notInChat, err := db.IsUserInChat(chat.ID, uuid.New())
	assert.Error(t, err)
	assert.False(t, notInChat)
}

func TestGetChat(t *testing.T) {
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

	fetchedChat, err := db.GetChat(chat.ID)
	assert.NoError(t, err)
	assert.Equal(t, chat, fetchedChat)

	_, err = db.GetChat(uuid.New())
	assert.Error(t, err)
}

func TestAddUserToChat(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user1 := &model.User{ID: uuid.New(), Username: "user1"}
	user2 := &model.User{ID: uuid.New(), Username: "user2"}
	_, err := db.CreateUser(user1)
	assert.NoError(t, err)
	_, err = db.CreateUser(user2)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user1.ID,
		Participants: []uuid.UUID{user1.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	err = db.AddUserToChat(chat.ID, user2.ID)
	assert.NoError(t, err)

	updatedChat, err := db.GetChat(chat.ID)
	assert.NoError(t, err)
	assert.Contains(t, updatedChat.Participants, user2.ID)

	err = db.AddUserToChat(uuid.New(), user2.ID)
	assert.Error(t, err)
}

func TestRemoveUserFromChat(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user1 := &model.User{ID: uuid.New(), Username: "user1"}
	user2 := &model.User{ID: uuid.New(), Username: "user2"}
	_, err := db.CreateUser(user1)
	assert.NoError(t, err)
	_, err = db.CreateUser(user2)
	assert.NoError(t, err)

	chat := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user1.ID,
		Participants: []uuid.UUID{user1.ID, user2.ID},
	}
	_, err = db.CreateChat(chat)
	assert.NoError(t, err)

	err = db.RemoveUserFromChat(chat.ID, user2.ID)
	assert.NoError(t, err)

	updatedChat, err := db.GetChat(chat.ID)
	assert.NoError(t, err)
	assert.NotContains(t, updatedChat.Participants, user2.ID)

	err = db.RemoveUserFromChat(uuid.New(), user2.ID)
	assert.Error(t, err)
}

func TestGetChatMessages(t *testing.T) {
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

	chatMessages, err := db.GetChatMessages(chat.ID)
	assert.NoError(t, err)
	assert.Len(t, chatMessages, 2)
	assert.Contains(t, chatMessages, message1)
	assert.Contains(t, chatMessages, message2)

	_, err = db.GetChatMessages(uuid.New())
	assert.Error(t, err)
}

func TestGetAllChats(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	chat1 := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePublic,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat1)
	assert.NoError(t, err)

	chat2 := &model.Chat{
		ID:           uuid.New(),
		Type:         model.ChatTypePrivate,
		CreatorID:    user.ID,
		Participants: []uuid.UUID{user.ID},
	}
	_, err = db.CreateChat(chat2)
	assert.NoError(t, err)

	chats, err := db.GetAllChats()
	assert.NoError(t, err)
	assert.Len(t, chats, 2)
	assert.Contains(t, chats, chat1)
	assert.Contains(t, chats, chat2)
}
