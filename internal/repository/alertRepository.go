package repository

import (
	"github.com/TradeLayers/BE/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AlertRepository interface {
	Create(db *gorm.DB, alert *model.Alert) error
	ListByUser(db *gorm.DB, userID string) ([]model.Alert, error)
	DeleteByUser(db *gorm.DB, userID string, alertID uuid.UUID) (bool, error)
	MarkTriggered(db *gorm.DB, alertID uuid.UUID) (*model.Alert, error)
}

type alertRepository struct{}

func NewAlertRepository() AlertRepository {
	return &alertRepository{}
}

func (r *alertRepository) Create(db *gorm.DB, alert *model.Alert) error {
	return db.Create(alert).Error
}

func (r *alertRepository) ListByUser(db *gorm.DB, userID string) ([]model.Alert, error) {
	var alerts []model.Alert = nil
	err := db.Where("user_id = ?", userID).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func (r *alertRepository) DeleteByUser(db *gorm.DB, userID string, alertID uuid.UUID) (bool, error) {
	result := db.Where("user_id = ? AND id = ?", userID, alertID).Delete(&model.Alert{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *alertRepository) MarkTriggered(db *gorm.DB, alertID uuid.UUID) (*model.Alert, error) {
	var alert model.Alert = model.Alert{}
	err := db.Raw(
		`UPDATE alerts
		 SET triggered_at = COALESCE(triggered_at, NOW())
		 WHERE id = ?
		 RETURNING id, user_id, stock_id, threshold_price, direction, triggered_at, created_at`,
		alertID,
	).Scan(&alert).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}
