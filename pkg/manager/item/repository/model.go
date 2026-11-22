package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"karat-server/common"
)

type Item struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID  `gorm:"type:uuid;index;not null"`
	Name        string     `gorm:"not null"`
	Kind        string     `gorm:"not null"`
	Metal       string     `gorm:"not null"`
	Purity      string     `gorm:"not null"`
	WeightGrams float64    `gorm:"not null"`
	CurrencyID  uuid.UUID  `gorm:"type:uuid;not null;index;default:'11111111-1111-4111-8111-111111111111'"`
	Cost        int64      `gorm:"not null"`
	Notes       string     `gorm:"not null;default:''"`
	ImageKey    string     `gorm:"column:image_key;not null;default:''"`
	CreatedOn   time.Time  `gorm:"type:timestamptz;not null"`
	SoldOn      *time.Time `gorm:"type:timestamptz"`
	SoldAmount  *int64
	common.CreatedUpdated
}

type Repository interface {
	Migrate() error
	List(ctx context.Context, userID uuid.UUID) ([]Item, error)
	Create(ctx context.Context, item *Item) error
	Get(ctx context.Context, userID, id uuid.UUID) (*Item, error)
	GetForUpdate(ctx context.Context, userID, id uuid.UUID) (*Item, error)
	Save(ctx context.Context, item *Item) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
	WithinTransaction(ctx context.Context, fn func(Repository) error) error
}
