package repository

import (
	"time"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(db *gorm.DB, notification *model.ThresholdNotification) error
	ListUnreadByUser(db *gorm.DB, userID string, limit int) ([]model.ThresholdNotification, error)
	MarkReadByUser(db *gorm.DB, userID string, notificationID uuid.UUID) (bool, error)
}

type notificationRepository struct{}

func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{}
}

func (r *notificationRepository) Create(db *gorm.DB, notification *model.ThresholdNotification) error {
	return db.Create(notification).Error
}

func (r *notificationRepository) ListUnreadByUser(db *gorm.DB, userID string, limit int) ([]model.ThresholdNotification, error) {
	if limit <= 0 {
		limit = 20
	}

	var notifications []model.ThresholdNotification
	err := db.Where("user_id = ? AND read_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error

	return notifications, err
}

func (r *notificationRepository) MarkReadByUser(db *gorm.DB, userID string, notificationID uuid.UUID) (bool, error) {
	now := time.Now().UTC()
	result := db.Model(&model.ThresholdNotification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", notificationID, userID).
		Update("read_at", now)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}