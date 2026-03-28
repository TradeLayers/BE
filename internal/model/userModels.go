package model

import (
	"time"

	"google.golang.org/genproto/googleapis/type/decimal"
)

type User struct {
	FirebaseId string          `gorm:"type:text;primaryKey" json:"firebaseId"`
	Name       string          `gorm:"type:varchar(255);not null" json:"name"`
	Email      string          `gorm:"type:varchar(255);not null" json:"email"`
	Balance    decimal.Decimal `gorm:"type:decimal(20,2);default:500.00" json:"balance"`
	CreatedAt  time.Time       `gorm:"type:timestamp;not null;autoCreateTime" json:"createdAt"`
	LastOnline time.Time       `gorm:"type:timestamp;not null;autoUpdateTime" json:"lastOnline"`
}

type UserContext struct {
	FirebaseId string
	Email      string
	Name       string
}

type UpdateFieldsDto struct {
	Email *string `json:"email"`
	Name  *string `json:"name"`
}

type FetchedOrCreated int

const (
	UserCreated FetchedOrCreated = iota
	UserFetched
	None
)

func (User) TableName() string {
	return "users"
}
