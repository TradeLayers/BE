package service

import (
	"context"
	"errors"
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"github.com/TradeLayers/BE/internal/requestlog"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService interface {
	CreateOrFetchUser(ctx context.Context, userCtx model.UserContext) (*model.User, model.FetchedOrCreated, appErrors.DomainError)
	UpdateFields(ctx context.Context, userCtx model.UserContext, fieldsObj model.UpdateFieldsDto) (*model.User, appErrors.DomainError)
	DeleteUser(ctx context.Context, userCtx model.UserContext) appErrors.DomainError
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) CreateOrFetchUser(ctx context.Context, userCtx model.UserContext) (*model.User, model.FetchedOrCreated, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	fetchedUser, err := s.repo.GetUser(ctx, userCtx)
	if err != nil {
		log.Error("failed to fetch user", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, model.None, appErrors.ErrInternal
	}

	if fetchedUser != nil {
		return fetchedUser, model.UserFetched, appErrors.ErrNone
	}

	createdUser, err := s.repo.CreateUser(ctx, userCtx)
	if err != nil {
		log.Error("failed to create user", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, model.None, appErrors.ErrInternal
	}

	return createdUser, model.UserCreated, appErrors.ErrNone
}

func (s *userService) UpdateFields(ctx context.Context, userCtx model.UserContext, fieldsObj model.UpdateFieldsDto) (*model.User, appErrors.DomainError) {
	log := requestlog.FromContext(ctx)

	if fieldsObj.Email == nil && fieldsObj.Name == nil {
		return nil, appErrors.ErrEmptyProvidedFields
	}

	updates := make(map[string]interface{})

	if fieldsObj.Email != nil {
		updates["email"] = *fieldsObj.Email
	}

	if fieldsObj.Name != nil {
		updates["name"] = *fieldsObj.Name
	}

	emailInvalid := fieldsObj.Email != nil && strings.TrimSpace(*fieldsObj.Email) == ""
	nameInvalid := fieldsObj.Name != nil && strings.TrimSpace(*fieldsObj.Name) == ""

	if emailInvalid && nameInvalid {
		return nil, appErrors.ErrEmptyProvidedFields
	}

	if emailInvalid || nameInvalid {
		return nil, appErrors.ErrInvalidFieldInformation
	}

	if err := s.repo.UpdateUser(ctx, userCtx, updates); err != nil {
		log.Error("failed to update user", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	user, err := s.repo.GetUser(ctx, userCtx)
	if err != nil {
		log.Error("failed to reload user after update", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return nil, appErrors.ErrInternal
	}

	if user == nil {
		return nil, appErrors.ErrUserNotFound
	}

	return user, appErrors.ErrNone
}

func (s *userService) DeleteUser(ctx context.Context, userCtx model.UserContext) appErrors.DomainError {
	err := s.repo.DeleteUser(ctx, userCtx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrUserNotFound
	}

	if err != nil {
		requestlog.FromContext(ctx).Error("failed to delete user", zap.String("firebase_id", userCtx.FirebaseId), zap.Error(err))
		return appErrors.ErrInternal
	}

	return appErrors.ErrNone
}
