package repository

import (
	"context"
	"time"

	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, db *gorm.DB, notification *model.ThresholdNotification) error
	ListUnreadByUser(ctx context.Context, db *gorm.DB, userID string, limit int) ([]model.ThresholdNotification, error)
	MarkReadByUser(ctx context.Context, db *gorm.DB, userID string, notificationID uuid.UUID) (bool, error)
}

type notificationRepository struct{}

func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{}
}

func (r *notificationRepository) Create(ctx context.Context, db *gorm.DB, notification *model.ThresholdNotification) error {
	return withContext(ctx, db).Create(notification).Error
}

func (r *notificationRepository) ListUnreadByUser(ctx context.Context, db *gorm.DB, userID string, limit int) ([]model.ThresholdNotification, error) {
	if limit <= 0 {
		limit = 20
	}

	var notifications []model.ThresholdNotification
	err := withContext(ctx, db).Where("user_id = ? AND read_at IS NULL", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error

	return notifications, err
}

func (r *notificationRepository) MarkReadByUser(ctx context.Context, db *gorm.DB, userID string, notificationID uuid.UUID) (bool, error) {
	now := time.Now().UTC()
	result := withContext(ctx, db).Model(&model.ThresholdNotification{}).
		Where("id = ? AND user_id = ? AND read_at IS NULL", notificationID, userID).
		Update("read_at", now)
	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}
