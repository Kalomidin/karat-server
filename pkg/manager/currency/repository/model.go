package repository

import "github.com/google/uuid"

type Currency struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code string    `gorm:"uniqueIndex;not null" json:"code"`
	Name string    `gorm:"not null" json:"name"`
}
