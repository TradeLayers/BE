package repository

import (
	"context"

	"github.com/TradeLayers/BE/internal/model"
)

type MockUserRepository struct {
	GetUserFn    func(ctx context.Context, userCtx model.UserContext) (*model.User, error)
	CreateUserFn func(ctx context.Context, userCtx model.UserContext) (*model.User, error)
	UpdateUserFn func(ctx context.Context, userCtx model.UserContext, updates map[string]interface{}) error
	DeleteUserFn func(ctx context.Context, userCtx model.UserContext) error
}

func (m *MockUserRepository) GetUser(ctx context.Context, userCtx model.UserContext) (*model.User, error) {
	return m.GetUserFn(ctx, userCtx)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, userCtx model.UserContext) (*model.User, error) {
	return m.CreateUserFn(ctx, userCtx)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, userCtx model.UserContext, updates map[string]interface{}) error {
	return m.UpdateUserFn(ctx, userCtx, updates)
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, userCtx model.UserContext) error {
	return m.DeleteUserFn(ctx, userCtx)
}
