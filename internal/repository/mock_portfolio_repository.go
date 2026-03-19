package repository

import (
	"context"

	"github.com/TradeLayers/BE/internal/model"
)

// MockPortfolioRepository is a test double for PortfolioRepository.
type MockPortfolioRepository struct {
	CreateFn    func(ctx context.Context, portfolio *model.Portfolio) error
	CreateCalls []*model.Portfolio
}

func (m *MockPortfolioRepository) Create(ctx context.Context, portfolio *model.Portfolio) error {
	m.CreateCalls = append(m.CreateCalls, portfolio)
	if m.CreateFn != nil {
		return m.CreateFn(ctx, portfolio)
	}
	return nil
}
