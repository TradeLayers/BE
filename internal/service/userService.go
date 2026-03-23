package service

import (
	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
)

type UserService interface {
	CreateOrFetchUser(userCtx model.UserContext) (*model.User, model.FetchedOrCreated, error)
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
