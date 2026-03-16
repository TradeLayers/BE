package model

import (
	"time"

	"github.com/google/uuid"
)

// Holding represents a position in an asset within a portfolio.
type Holding struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	PortfolioID uuid.UUID `gorm:"type:uuid;not null;index"`
	Symbol      string    `gorm:"type:varchar(50);not null"`
	Quantity    float64   `gorm:"type:decimal(20,8);not null"`
	AverageCost float64   `gorm:"column:average_cost;type:decimal(20,2);not null"`
	CreatedAt   time.Time `gorm:"type:timestamp;not null"`
	UpdatedAt   time.Time `gorm:"type:timestamp;not null"`
}

// TableName overrides the table name to "holdings".
func (Holding) TableName() string {
	return "holdings"
}
