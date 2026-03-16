package model

import (
	"time"

	"github.com/google/uuid"
)

// Portfolio represents an investment portfolio.
type Portfolio struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name           string    `gorm:"type:varchar(255);not null"`
	Balance        float64   `gorm:"type:decimal(20,2);not null"`
	InitialBalance float64   `gorm:"column:initial_balance;type:decimal(20,2);not null"`
	CreatedAt      time.Time `gorm:"type:timestamp;not null"`
	UpdatedAt      time.Time `gorm:"type:timestamp;not null"`
}

// TableName overrides the table name to "portfolios".
func (Portfolio) TableName() string {
	return "portfolios"
}

