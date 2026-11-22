package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"karat-server/common"
	currency_repo "karat-server/pkg/manager/currency/repository"
)

type repository struct {
	*gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) Migrate() error {
	if err := r.AutoMigrate(&Item{}); err != nil {
		return err
	}
	itemTable, err := tableName(r.DB, &Item{})
	if err != nil {
		return err
	}
	currencyTable, err := tableName(r.DB, &currency_repo.Currency{})
	if err != nil {
		return err
	}
	return r.Exec(`
DO $$ BEGIN
  ALTER TABLE ` + itemTable + `
    ADD CONSTRAINT item_currency_id_fkey
    FOREIGN KEY (currency_id) REFERENCES ` + currencyTable + ` (id);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;`).Error
}

func tableName(db *gorm.DB, model interface{}) (string, error) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return "", err
	}
	return stmt.Schema.Table, nil
}

func (r *repository) List(ctx context.Context, userID uuid.UUID) ([]Item, error) {
	var items []Item
	err := r.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC, id DESC").
		Find(&items).Error
	return items, err
}

func (r *repository) Create(ctx context.Context, item *Item) error {
	return r.WithContext(ctx).Create(item).Error
}

func (r *repository) Get(ctx context.Context, userID, id uuid.UUID) (*Item, error) {
	return r.one(ctx, userID, id, false)
}

func (r *repository) GetForUpdate(ctx context.Context, userID, id uuid.UUID) (*Item, error) {
	return r.one(ctx, userID, id, true)
}

func (r *repository) Save(ctx context.Context, item *Item) error {
	return r.WithContext(ctx).Select("*").Save(item).Error
}

func (r *repository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&Item{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrMissing
	}
	return nil
}

func (r *repository) WithinTransaction(ctx context.Context, fn func(Repository) error) error {
	return r.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&repository{tx})
	})
}

func (r *repository) one(ctx context.Context, userID, id uuid.UUID, lock bool) (*Item, error) {
	query := r.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID)
	if lock {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var item Item
	err := query.First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, common.ErrMissing
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}
