package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"messenger/internal/handler"
	"messenger/internal/model"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(username string) (*model.User, error) {
	args := m.Called(username)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetUser(requesterID, id uuid.UUID) (*model.User, error) {
	args := m.Called(requesterID, id)
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]*model.User, error) {
	args := m.Called()
	return args.Get(0).([]*model.User), args.Error(1)
}

func TestCreateUser(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	user := &model.User{
		ID:       uuid.New(),
		Username: "тестовый_пользователь",
	}

	mockService.On("CreateUser", "тестовый_пользователь").Return(user, nil)

	reqBody, _ := json.Marshal(map[string]string{"username": "тестовый_пользователь"})
	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
	rr := httptest.NewRecorder()

	http.HandlerFunc(userHandler.CreateUser).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Username, response.Username)

	mockService.AssertExpectations(t)
}

func TestGetUser(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	userID := uuid.New()
	requesterID := uuid.New()
	user := &model.User{
		ID:       userID,
		Username: "тестовый_пользователь",
	}

	mockService.On("GetUser", requesterID, userID).Return(user, nil)

	req, _ := http.NewRequest("GET", "/users/id="+userID.String(), nil)
	req.Header.Set("User-ID", requesterID.String()) // Имитируем заголовок с ID запрашивающего пользователя
	rr := httptest.NewRecorder()

	ctx := req.Context()
	ctx = context.WithValue(ctx, "userID", requesterID)
	req = req.WithContext(ctx)

	http.HandlerFunc(userHandler.GetUser).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response model.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, response.ID)
	assert.Equal(t, user.Username, response.Username)

	mockService.AssertExpectations(t)
}

func TestGetAllUsers(t *testing.T) {
	mockService := new(MockUserService)
	userHandler := handler.NewUserHandler(mockService)

	users := []*model.User{
		{ID: uuid.New(), Username: "пользователь1"},
		{ID: uuid.New(), Username: "пользователь2"},
	}

	mockService.On("GetAllUsers").Return(users, nil)

	_, _ = http.NewRequest("GET", "/users", nil)
	rr := httptest.NewRecorder()

	// Вызываем метод напрямую, передавая ResponseWriter
	userHandler.GetAllUsers(rr)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response []*model.User
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, len(users), len(response))
	for i, user := range users {
		assert.Equal(t, user.ID, response[i].ID)
		assert.Equal(t, user.Username, response[i].Username)
	}

	mockService.AssertExpectations(t)
}
