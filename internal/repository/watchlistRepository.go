package repository

import (
	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WatchlistRepository interface {
	ListByUser(db *gorm.DB, userID string) ([]model.WatchlistEntry, error)
	Add(db *gorm.DB, userID string, stockID uuid.UUID) error
	Remove(db *gorm.DB, userID string, stockID uuid.UUID) (bool, error)
	Exists(db *gorm.DB, userID string, stockID uuid.UUID) (bool, error)
}

type watchlistRepository struct{}

func NewWatchlistRepository() WatchlistRepository {
	return &watchlistRepository{}
}

func (r *watchlistRepository) ListByUser(db *gorm.DB, userID string) ([]model.WatchlistEntry, error) {
	var entries []model.WatchlistEntry
	err := db.Where("user_id = ?", userID).Order("added_at DESC").Find(&entries).Error
	return entries, err
}

func (r *watchlistRepository) Add(db *gorm.DB, userID string, stockID uuid.UUID) error {
	return db.Create(&model.WatchlistEntry{UserID: userID, StockID: stockID}).Error
}

func (r *watchlistRepository) Remove(db *gorm.DB, userID string, stockID uuid.UUID) (bool, error) {
	result := db.Where("user_id = ? AND stock_id = ?", userID, stockID).
		Delete(&model.WatchlistEntry{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *watchlistRepository) Exists(db *gorm.DB, userID string, stockID uuid.UUID) (bool, error) {
	var count int64
	err := db.Model(&model.WatchlistEntry{}).
		Where("user_id = ? AND stock_id = ?", userID, stockID).
		Count(&count).Error
	return count > 0, err
}
