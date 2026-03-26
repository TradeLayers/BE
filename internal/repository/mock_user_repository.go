package repository

import "github.com/TradeLayers/BE/internal/model"

type MockUserRepository struct {
	GetUserFn    func(userCtx model.UserContext) (*model.User, error)
	CreateUserFn func(userCtx model.UserContext) (*model.User, error)
	UpdateUserFn func(userCtx model.UserContext, updates map[string]interface{}) error
	DeleteUserFn func(userCtx model.UserContext) error
}

func (m *MockUserRepository) GetUser(userCtx model.UserContext) (*model.User, error) {
	return m.GetUserFn(userCtx)
}

func (m *MockUserRepository) CreateUser(userCtx model.UserContext) (*model.User, error) {
	return m.CreateUserFn(userCtx)
}

func (m *MockUserRepository) UpdateUser(userCtx model.UserContext, updates map[string]interface{}) error {
	return m.UpdateUserFn(userCtx, updates)
}

func (m *MockUserRepository) DeleteUser(userCtx model.UserContext) error {
	return m.DeleteUserFn(userCtx)
}
