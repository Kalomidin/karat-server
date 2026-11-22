package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"karat-server/common"
	"karat-server/pkg/currency"
)

type Repository interface {
	Migrate() error
	List(ctx context.Context) ([]Currency, error)
	Require(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	*gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Migrate() error {
	if err := r.AutoMigrate(&Currency{}); err != nil {
		return err
	}
	for _, seed := range currency.Catalog() {
		var existing Currency
		err := r.Where("code = ?", seed.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row := Currency{ID: seed.ID, Code: seed.Code, Name: seed.Name}
			if err := r.Create(&row).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) List(ctx context.Context) ([]Currency, error) {
	var rows []Currency
	err := r.WithContext(ctx).Order("code").Find(&rows).Error
	return rows, err
}

func (r *repository) Require(ctx context.Context, id uuid.UUID) error {
	var row Currency
	err := r.WithContext(ctx).Where("id = ?", id).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return common.ErrInvalidCurrency
	}
	return err
}
