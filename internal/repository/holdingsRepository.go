package repository

import (
	"context"
	"errors"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type HoldingsRepository interface {
	GetByUser(ctx context.Context, db *gorm.DB, userID string) ([]model.UserHoldings, error)
	GetOne(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) (*model.UserHoldings, error)
	Upsert(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, delta decimal.Decimal) error
	SetQuantity(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, qty decimal.Decimal) error
	Delete(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) error
}

type holdingsRepository struct{}

func NewHoldingsRepository() HoldingsRepository {
	return &holdingsRepository{}
}

func (r *holdingsRepository) GetByUser(ctx context.Context, db *gorm.DB, userID string) ([]model.UserHoldings, error) {
	var holdings []model.UserHoldings
	err := withContext(ctx, db).Where("user_id = ?", userID).Find(&holdings).Error
	return holdings, err
}

func (r *holdingsRepository) GetOne(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) (*model.UserHoldings, error) {
	var h model.UserHoldings
	err := withContext(ctx, db).Where("user_id = ? AND stock_id = ?", userID, stockID).First(&h).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *holdingsRepository) Upsert(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, delta decimal.Decimal) error {
	existing, err := r.GetOne(ctx, db, userID, stockID)
	if err != nil {
		return err
	}
	if existing == nil {
		return withContext(ctx, db).Create(&model.UserHoldings{
			UserID:   userID,
			StockID:  stockID,
			Quantity: delta,
		}).Error
	}
	newQty := existing.Quantity.Add(delta)
	return withContext(ctx, db).Model(&model.UserHoldings{}).
		Where("user_id = ? AND stock_id = ?", userID, stockID).
		Update("quantity", newQty).Error
}

func (r *holdingsRepository) SetQuantity(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, qty decimal.Decimal) error {
	return withContext(ctx, db).Model(&model.UserHoldings{}).
		Where("user_id = ? AND stock_id = ?", userID, stockID).
		Update("quantity", qty).Error
}

func (r *holdingsRepository) Delete(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) error {
	return withContext(ctx, db).Where("user_id = ? AND stock_id = ?", userID, stockID).
		Delete(&model.UserHoldings{}).Error
}
