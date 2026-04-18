package model

import (
	"time"

	"github.com/google/uuid"
)

type WatchlistEntry struct {
	UserID  string    `gorm:"column:user_id;type:text;primaryKey" json:"userId"`
	StockID uuid.UUID `gorm:"column:stock_id;type:uuid;primaryKey" json:"stockId"`
	AddedAt time.Time `gorm:"column:added_at;type:timestamp;not null;autoCreateTime" json:"addedAt"`
}

func (WatchlistEntry) TableName() string {
	return "watchlist"
}
