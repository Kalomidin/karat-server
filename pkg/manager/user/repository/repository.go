package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type repository struct {
	*gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Migrate() error {
	return r.AutoMigrate(&User{})
}

func (r *repository) UpsertByGoogleID(ctx context.Context, user *User) (*User, error) {
	var existing User
	err := r.WithContext(ctx).Where("google_id = ?", user.GoogleID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.WithContext(ctx).Create(user).Error; err != nil {
			return nil, err
		}
		return user, nil
	}
	if err != nil {
		return nil, err
	}
	existing.Email = user.Email
	existing.Name = user.Name
	existing.PhotoURL = user.PhotoURL
	if err := r.WithContext(ctx).Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}
