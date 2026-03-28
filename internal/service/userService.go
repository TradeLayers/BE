package service

import (
	"errors"
	"strings"

	"github.com/TradeLayers/BE/internal/appErrors"
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
	"gorm.io/gorm"
)

type UserService interface {
	CreateOrFetchUser(userCtx model.UserContext) (*model.User, model.FetchedOrCreated, appErrors.DomainError)
	UpdateFields(userCtx model.UserContext, fieldsObj model.UpdateFieldsDto) (*model.User, appErrors.DomainError)
	DeleteUser(userCtx model.UserContext) appErrors.DomainError
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) CreateOrFetchUser(userCtx model.UserContext) (*model.User, model.FetchedOrCreated, appErrors.DomainError) {
	fetchedUser, err := s.repo.GetUser(userCtx)
	if err != nil {
		return nil, model.None, appErrors.ErrInternal
	}

	if fetchedUser != nil {
		return fetchedUser, model.UserFetched, appErrors.ErrNone
	}

	createdUser, err := s.repo.CreateUser(userCtx)
	if err != nil {
		return nil, model.None, appErrors.ErrInternal
	}

	return createdUser, model.UserCreated, appErrors.ErrNone
}

func (s *userService) UpdateFields(userCtx model.UserContext, fieldsObj model.UpdateFieldsDto) (*model.User, appErrors.DomainError) {

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

	if err := s.repo.UpdateUser(userCtx, updates); err != nil {
		return nil, appErrors.ErrInternal
	}

	user, err := s.repo.GetUser(userCtx)
	if err != nil {
		return nil, appErrors.ErrInternal
	}

	if user == nil {
		return nil, appErrors.ErrUserNotFound
	}

	return user, appErrors.ErrNone
}

func (s *userService) DeleteUser(userCtx model.UserContext) appErrors.DomainError {
	err := s.repo.DeleteUser(userCtx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appErrors.ErrUserNotFound
	}

	if err != nil {
		return appErrors.ErrInternal
	}

	return appErrors.ErrNone
}
