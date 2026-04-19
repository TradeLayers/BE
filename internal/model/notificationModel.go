package model

import (
	"time"

	"github.com/google/uuid"
)

type ThresholdNotification struct {
	ID             uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID         string     `gorm:"column:user_id;type:text;not null" json:"userId"`
	StockID        uuid.UUID  `gorm:"column:stock_id;type:uuid;not null" json:"stockId"`
	Symbol         string     `gorm:"column:symbol;type:text;not null" json:"symbol"`
	ThresholdPrice float64    `gorm:"column:threshold_price;type:double precision;not null" json:"thresholdPrice"`
	TriggerPrice   float64    `gorm:"column:trigger_price;type:double precision;not null" json:"triggerPrice"`
	CreatedAt      time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime" json:"createdAt"`
	ReadAt         *time.Time `gorm:"column:read_at;type:timestamp" json:"readAt"`
}

func (ThresholdNotification) TableName() string {
	return "threshold_notifications"
}