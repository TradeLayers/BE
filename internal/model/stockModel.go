package model

import "github.com/google/uuid"

type Stock struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	StockName string    `gorm:"column:stock_name;type:varchar(128);not null"`
	Symbol    string    `gorm:"type:varchar(32);not null;unique"`
}

func (Stock) TableName() string {
	return "stocks"
}
