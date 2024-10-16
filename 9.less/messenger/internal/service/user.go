package service

import (
	"errors"
	"github.com/google/uuid"
	"messenger/internal/model"
	"messenger/internal/storage"
)

type UserService struct {
	storage  storage.Storage
	maxUsers int
}

func NewUserService(storage storage.Storage, maxUsers int) *UserService {
	return &UserService{storage: storage, maxUsers: maxUsers}
}

func (s *UserService) CreateUser(username string) (*model.User, error) {
	if s.storage.GetUserCount() >= s.maxUsers {
		return nil, errors.New("maximum number of users reached")
	}

	user := &model.User{
		ID:       uuid.New(),
		Username: username,
	}
	return s.storage.CreateUser(user)
}

func (s *UserService) GetUser(id uuid.UUID) (*model.User, error) {
	return s.storage.GetUser(id)
}

func (s *UserService) GetAllUsers() ([]*model.User, error) {
	return s.storage.GetAllUsers()
}
