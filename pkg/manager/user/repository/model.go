package repository

import (
	"context"

	"github.com/google/uuid"

	"karat-server/common"
)

type User struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	GoogleID string    `gorm:"column:google_id;uniqueIndex;not null"`
	Email    string    `gorm:"not null"`
	Name     string    `gorm:"not null;default:''"`
	PhotoURL string    `gorm:"column:photo_url;not null;default:''"`
	common.CreatedUpdated
}

type Repository interface {
	Migrate() error
	UpsertByGoogleID(ctx context.Context, user *User) (*User, error)
}
