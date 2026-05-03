package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type AlertDirection string

const (
	AlertDirectionAbove AlertDirection = "above"
	AlertDirectionBelow AlertDirection = "below"
)

type Alert struct {
	ID             uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID         string          `gorm:"column:user_id;type:text;not null"`
	StockID        uuid.UUID       `gorm:"column:stock_id;type:uuid;not null"`
	ThresholdPrice decimal.Decimal `gorm:"column:threshold_price;type:decimal(20,2);not null"`
	Direction      AlertDirection  `gorm:"column:direction;type:alert_direction;not null"`
	TriggeredAt    *time.Time      `gorm:"column:triggered_at;type:timestamp"`
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamp;not null;autoCreateTime"`
}

func (Alert) TableName() string {
	return "alerts"
}
