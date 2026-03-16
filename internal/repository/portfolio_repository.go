package repository

import (
	"context"

	"github.com/TradeLayers/BE/internal/model"
	"gorm.io/gorm"
)

type PortfolioRepository interface {
	Create(ctx context.Context, portfolio *model.Portfolio) error
}

type portfolioRepository struct {
	db *gorm.DB
}

func NewPortfolioRepository(db *gorm.DB) PortfolioRepository {
	return &portfolioRepository{db: db}
}

func (r *portfolioRepository) Create(ctx context.Context, portfolio *model.Portfolio) error {
	return r.db.WithContext(ctx).Create(portfolio).Error
}

