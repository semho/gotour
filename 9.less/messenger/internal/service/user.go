package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"messenger/internal/model"
	"messenger/internal/storage"
)

type userService struct {
	storage  storage.Storage
	maxUsers int
}

func NewUserService(storage storage.Storage, maxUsers int) UserService {
	return &userService{storage: storage, maxUsers: maxUsers}
}

func (s *userService) CreateUser(username string) (*model.User, error) {
	if s.storage.GetUserCount() >= s.maxUsers {
		return nil, errors.New("maximum number of users reached")
	}

	user := &model.User{
		ID:       uuid.New(),
		Username: username,
	}
	return s.storage.CreateUser(user)
}

func (s *userService) GetUser(requesterID, id uuid.UUID) (*model.User, error) {
	_, err := s.storage.GetUser(requesterID)
	if err != nil {
		return nil, fmt.Errorf("requester not found: %w", err)
	}
	return s.storage.GetUser(id)
}

func (s *userService) GetAllUsers() ([]*model.User, error) {
	return s.storage.GetAllUsers()
}
