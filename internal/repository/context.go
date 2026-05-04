package repository

import (
	"context"

	"gorm.io/gorm"
)

func withContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if ctx == nil {
		return db
	}
	return db.WithContext(ctx)
}
