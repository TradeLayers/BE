package repository

import (
	"errors"

	"github.com/TradeLayers/BE/internal/model"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUser(userCtx model.UserContext) (*model.User, error)
	CreateUser(userCtx model.UserContext) (*model.User, error)
	UpdateUser(userCtx model.UserContext, updates map[string]interface{}) error
	DeleteUser(userCtx model.UserContext) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetUser(userCtx model.UserContext) (*model.User, error) {
	var user model.User

	err := r.db.Where("firebase_id = ?", userCtx.FirebaseId).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) CreateUser(userCtx model.UserContext) (*model.User, error) {
	user := model.User{
		FirebaseId: userCtx.FirebaseId,
		Name:       userCtx.Name,
		Email:      userCtx.Email,
	}

	err := r.db.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpdateUser(userCtx model.UserContext, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}

	err := r.db.Model(&model.User{}).
		Where("firebase_id = ?", userCtx.FirebaseId).
		Updates(updates).Error

	return err
}

func (r *userRepository) DeleteUser(userCtx model.UserContext) error {
	err := r.db.Where("firebase_id = ?", userCtx.FirebaseId).Delete(&model.User{}).Error
	return err
}
