package model

import (
	"time"

	"github.com/google/uuid"
)

type WatchlistEntry struct {
	UserID         string    `gorm:"column:user_id;type:text;primaryKey" json:"userId"`
	StockID        uuid.UUID `gorm:"column:stock_id;type:uuid;primaryKey" json:"stockId"`
	AddedAt        time.Time `gorm:"column:added_at;type:timestamp;not null;autoCreateTime" json:"addedAt"`
	ThresholdPrice *float64  `gorm:"column:threshold_price;type:double precision" json:"thresholdPrice"`
}

func (WatchlistEntry) TableName() string {
	return "watchlist"
}
