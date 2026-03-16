package service

import (
	"context"
	"time"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/TradeLayers/BE/internal/repository"
)

type PortfolioService interface {
	CreatePortfolio(ctx context.Context, name string, initialBalance float64) (*model.Portfolio, error)
}

type portfolioService struct {
	repo repository.PortfolioRepository
}

func NewPortfolioService(repo repository.PortfolioRepository) PortfolioService {
	return &portfolioService{repo: repo}
}

func (s *portfolioService) CreatePortfolio(ctx context.Context, name string, initialBalance float64) (*model.Portfolio, error) {
	now := time.Now()

	p := &model.Portfolio{
		Name:           name,
		Balance:        initialBalance,
		InitialBalance: initialBalance,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

