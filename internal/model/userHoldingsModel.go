package model

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UserHoldings struct {
	UserID   string          `gorm:"column:user_id;type:text;primaryKey" json:"userId"`
	StockID  uuid.UUID       `gorm:"column:stock_id;type:uuid;primaryKey" json:"stockId"`
	Quantity decimal.Decimal `gorm:"type:decimal(20,2);not null" json:"quantity"`
}

func (UserHoldings) TableName() string {
	return "users_holdings"
}
