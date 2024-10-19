package tests

import (
	"messenger/internal/model"
	"messenger/internal/storage/memory"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
	}

	createdUser, err := db.CreateUser(user)
	assert.NoError(t, err)
	assert.Equal(t, user, createdUser)

	fetchedUser, err := db.GetUser(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user, fetchedUser)
}

func TestGetUserCount(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	assert.Equal(t, 0, db.GetUserCount())

	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	assert.Equal(t, 1, db.GetUserCount())
}

func TestGetUser(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user := &model.User{ID: uuid.New(), Username: "testuser"}
	_, err := db.CreateUser(user)
	assert.NoError(t, err)

	fetchedUser, err := db.GetUser(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user, fetchedUser)

	_, err = db.GetUser(uuid.New())
	assert.Error(t, err)
}

func TestGetAllUsers(t *testing.T) {
	db := memory.NewDB(100, 1000, 100)
	user1 := &model.User{ID: uuid.New(), Username: "user1"}
	user2 := &model.User{ID: uuid.New(), Username: "user2"}

	_, err := db.CreateUser(user1)
	assert.NoError(t, err)
	_, err = db.CreateUser(user2)
	assert.NoError(t, err)

	users, err := db.GetAllUsers()
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Contains(t, users, user1)
	assert.Contains(t, users, user2)
}
