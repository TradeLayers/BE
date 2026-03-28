package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionType string

const (
	TransactionTypeBought TransactionType = "bought"
	TransactionTypeSold   TransactionType = "sold"
)

type StockTransaction struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID          string          `gorm:"column:user_id;type:text;not null"`
	StockID         uuid.UUID       `gorm:"column:stock_id;type:uuid;not null"`
	Price           decimal.Decimal `gorm:"type:decimal(20,2);not null"`
	Quantity        decimal.Decimal `gorm:"type:decimal(20,2);not null"`
	TransactionDate time.Time       `gorm:"column:transaction_date;type:timestamp;not null;default:now();autoCreateTime"`
	TransactionType TransactionType `gorm:"column:transaction_type;type:transaction_type;not null"`
}

func (StockTransaction) TableName() string {
	return "stock_transactions"
}
