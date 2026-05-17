package repository

import (
	"context"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WatchlistRepository interface {
	ListByUser(ctx context.Context, db *gorm.DB, userID string) ([]model.WatchlistEntry, error)
	Add(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) error
	Remove(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) (bool, error)
	Exists(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) (bool, error)
	UpdateThreshold(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, thresholdPrice float64) (bool, error)
	UpdateThresholdReached(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, reached bool) (bool, error)
}

type watchlistRepository struct{}

func NewWatchlistRepository() WatchlistRepository {
	return &watchlistRepository{}
}

func (r *watchlistRepository) ListByUser(ctx context.Context, db *gorm.DB, userID string) ([]model.WatchlistEntry, error) {
	var entries []model.WatchlistEntry = nil
	err := withContext(ctx, db).Where("user_id = ?", userID).Order("added_at DESC").Find(&entries).Error
	return entries, err
}

func (r *watchlistRepository) Add(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) error {
	return withContext(ctx, db).Create(&model.WatchlistEntry{UserID: userID, StockID: stockID}).Error
}

func (r *watchlistRepository) Remove(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) (bool, error) {
	result := withContext(ctx, db).Where("user_id = ? AND stock_id = ?", userID, stockID).
		Delete(&model.WatchlistEntry{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *watchlistRepository) Exists(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID) (bool, error) {
	var count int64 = 0
	err := withContext(ctx, db).Model(&model.WatchlistEntry{}).
		Where("user_id = ? AND stock_id = ?", userID, stockID).
		Count(&count).Error
	return count > 0, err
}

func (r *watchlistRepository) UpdateThreshold(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, thresholdPrice float64) (bool, error) {
	result := withContext(ctx, db).Model(&model.WatchlistEntry{}).
		Where("user_id = ? AND stock_id = ?", userID, stockID).
		Updates(map[string]interface{}{
			"threshold_price":   thresholdPrice,
			"threshold_reached": false,
		})
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

func (r *watchlistRepository) UpdateThresholdReached(ctx context.Context, db *gorm.DB, userID string, stockID uuid.UUID, reached bool) (bool, error) {
	result := withContext(ctx, db).Model(&model.WatchlistEntry{}).
		Where("user_id = ? AND stock_id = ?", userID, stockID).
		Update("threshold_reached", reached)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
