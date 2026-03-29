package repository

import (
	"errors"

	"github.com/TradeLayers/BE/internal/model"
	"gorm.io/gorm"
)

type StockRepository interface {
	GetBySymbol(symbol string) (*model.Stock, error)
	GetAll() ([]model.Stock, error)
	Create(stock *model.Stock) error
}

type stockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) StockRepository {
	return &stockRepository{db: db}
}

func (r *stockRepository) GetBySymbol(symbol string) (*model.Stock, error) {
	var stock model.Stock

	err := r.db.Where("symbol = ?", symbol).First(&stock).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &stock, nil
}

func (r *stockRepository) GetAll() ([]model.Stock, error) {
	var stocks []model.Stock

	err := r.db.Find(&stocks).Error
	if err != nil {
		return nil, err
	}

	return stocks, nil
}

func (r *stockRepository) Create(stock *model.Stock) error {
	return r.db.Create(stock).Error
}
