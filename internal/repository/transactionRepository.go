package repository

import (
	"time"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionsRepository interface {
	Create(db *gorm.DB, tx *model.StockTransaction) error
	ListByUser(db *gorm.DB, userID string, stockID *uuid.UUID, from, to *time.Time) ([]model.StockTransaction, error)
}

type transactionsRepository struct{}

func NewTransactionsRepository() TransactionsRepository {
	return &transactionsRepository{}
}

func (r *transactionsRepository) Create(db *gorm.DB, tx *model.StockTransaction) error {
	return db.Create(tx).Error
}

func (r *transactionsRepository) ListByUser(db *gorm.DB, userID string, stockID *uuid.UUID, from, to *time.Time) ([]model.StockTransaction, error) {
	var txs []model.StockTransaction
	q := db.Where("user_id = ?", userID)
	if stockID != nil {
		q = q.Where("stock_id = ?", *stockID)
	}
	if from != nil {
		q = q.Where("transaction_date >= ?", *from)
	}
	if to != nil {
		q = q.Where("transaction_date <= ?", *to)
	}
	err := q.Order("transaction_date ASC").Find(&txs).Error
	return txs, err
}
