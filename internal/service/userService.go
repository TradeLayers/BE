package service

import (
	"strings"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
)

type UserService interface {
	CreateOrFetchUser(userCtx model.UserContext) (*model.User, model.FetchedOrCreated, error)
	UpdateFields(userCtx model.UserContext, fieldsObj model.UpdateFieldsDto) (*model.User, error)
	DeleteUser(userCtx model.UserContext) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) CreateOrFetchUser(userCtx model.UserContext) (*model.User, model.FetchedOrCreated, error) {
	fetchedUser, err := s.repo.GetUser(userCtx)

	if err != nil {
		return nil, model.None, err
	}

	if fetchedUser != nil {
		return fetchedUser, model.UserFetched, nil
	}

	createdUser, err := s.repo.CreateUser(userCtx)
	if err != nil {
		return nil, model.None, err
	}

	return createdUser, model.UserCreated, nil
}

func (s *userService) UpdateFields(userCtx model.UserContext, fieldsObj model.UpdateFieldsDto) (*model.User, error) {
	updates := make(map[string]interface{})

	if fieldsObj.Email != nil && strings.TrimSpace(*fieldsObj.Email) != "" {
		updates["email"] = *fieldsObj.Email
	}

	if fieldsObj.Name != nil && strings.TrimSpace(*fieldsObj.Name) != "" {
		updates["name"] = *fieldsObj.Name
	}

	err := s.repo.UpdateUser(userCtx, updates)

	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUser(userCtx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) DeleteUser(userCtx model.UserContext) error {
	err := s.repo.DeleteUser(userCtx)
	return err
}
