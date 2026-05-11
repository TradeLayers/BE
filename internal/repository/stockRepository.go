package repository

import (
	"context"
	"errors"

	"github.com/TradeLayers/BE/internal/model"
	"gorm.io/gorm"
)

type StockRepository interface {
	GetBySymbol(ctx context.Context, symbol string) (*model.Stock, error)
	GetAll(ctx context.Context) ([]model.Stock, error)
	Create(ctx context.Context, stock *model.Stock) error
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) GetBySymbol(ctx context.Context, symbol string) (*model.Stock, error) {
	var stock model.Stock = model.Stock{}

	err := withContext(ctx, r.db).Where("symbol = ?", symbol).First(&stock).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &stock, nil
}

func (r *stockRepository) GetAll(ctx context.Context) ([]model.Stock, error) {
	var stocks []model.Stock = nil

	err := withContext(ctx, r.db).Find(&stocks).Error
	if err != nil {
		return nil, err
	}

	return stocks, nil
}

func (r *stockRepository) Create(ctx context.Context, stock *model.Stock) error {
	return withContext(ctx, r.db).Create(stock).Error
}
