package memory

import (
	"errors"
	"github.com/google/uuid"
	"messenger/internal/model"
)

func (db *DB) CreateUser(user *model.User) (*model.User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.users[user.ID]; exists {
		return nil, errors.New("user already exists")
	}

	db.users[user.ID] = user
	return user, nil
}

func (db *DB) GetUserCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.users)
}

func (db *DB) GetUser(id uuid.UUID) (*model.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (db *DB) GetAllUsers() ([]*model.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	users := make([]*model.User, 0, len(db.users))
	for _, user := range db.users {
		users = append(users, user)
	}

	return users, nil
}
