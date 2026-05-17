package repository

import (
	"context"
	"errors"

	"github.com/TradeLayers/BE/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUser(ctx context.Context, userCtx model.UserContext) (*model.User, error)
	CreateUser(ctx context.Context, userCtx model.UserContext) (*model.User, error)
	UpdateUser(ctx context.Context, userCtx model.UserContext, updates map[string]interface{}) error
	DeleteUser(ctx context.Context, userCtx model.UserContext) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUser(ctx context.Context, userCtx model.UserContext) (*model.User, error) {
	var user model.User = model.User{}

	err := withContext(ctx, r.db).Where("firebase_id = ?", userCtx.FirebaseId).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, userCtx model.UserContext) (*model.User, error) {
	user := model.User{
		FirebaseId: userCtx.FirebaseId,
		Name:       userCtx.Name,
		Email:      userCtx.Email,
	}

	err := withContext(ctx, r.db).Create(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, userCtx model.UserContext, updates map[string]interface{}) error {
	err := withContext(ctx, r.db).Model(&model.User{}).Where("firebase_id = ?", userCtx.FirebaseId).Updates(updates).Error
	return err
}

func (r *userRepository) DeleteUser(ctx context.Context, userCtx model.UserContext) error {
	result := withContext(ctx, r.db).
		Where("firebase_id = ?", userCtx.FirebaseId).
		Delete(&model.User{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
