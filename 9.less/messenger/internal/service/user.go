package service

import (
	"github.com/google/uuid"
	"messenger/internal/model"
	"messenger/internal/storage/memory"
)

type UserService struct {
	storage *memory.DB
}

func NewUserService(storage *memory.DB) *UserService {
	return &UserService{storage: storage}
}

func (s *UserService) CreateUser(username string) (*model.User, error) {
	user := &model.User{
		ID:       uuid.New(),
		Username: username,
	}
	return s.storage.CreateUser(user)
}

func (s *UserService) GetUser(id uuid.UUID) (*model.User, error) {
	return s.storage.GetUser(id)
}
