package repository

import (
	"context"
	"errors"
	"strings"

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
	if err := utcTimestamps(r.DB, itemTable); err != nil {
		return err
	}
	if err := dropMetalCatalog(r.DB, itemTable); err != nil {
		return err
	}
	currencyTable, err := tableName(r.DB, &currency_repo.Currency{})
	if err != nil {
		return err
	}
	return addForeignKey(r.DB, itemTable, "item_currency_id_fkey", "currency_id", currencyTable)
}

func dropMetalCatalog(db *gorm.DB, itemTable string) error {
	schema, _, ok := strings.Cut(itemTable, ".")
	if !ok {
		schema = "public"
	}
	if err := db.Exec(`ALTER TABLE ` + itemTable + ` DROP COLUMN IF EXISTS metal_id`).Error; err != nil {
		return err
	}
	return db.Exec(`DROP TABLE IF EXISTS ` + schema + `.metal`).Error
}

func addForeignKey(db *gorm.DB, table, name, column, refTable string) error {
	return db.Exec(`
DO $$ BEGIN
  ALTER TABLE ` + table + `
    ADD CONSTRAINT ` + name + `
    FOREIGN KEY (` + column + `) REFERENCES ` + refTable + ` (id);
EXCEPTION
  WHEN duplicate_object THEN NULL;
END $$;`).Error
}

func utcTimestamps(db *gorm.DB, table string) error {
	return db.Exec(`
DO $$ BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = split_part('` + table + `', '.', 1)
      AND table_name = split_part('` + table + `', '.', 2)
      AND column_name = 'created_on'
      AND data_type = 'date'
  ) THEN
    ALTER TABLE ` + table + `
      ALTER COLUMN created_on TYPE timestamptz USING created_on::timestamp AT TIME ZONE 'UTC',
      ALTER COLUMN sold_on TYPE timestamptz USING sold_on::timestamp AT TIME ZONE 'UTC';
  END IF;
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
		Where("user_id = ? AND deleted_at IS NULL", userID).
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
	res := r.WithContext(ctx).
		Model(&Item{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", item.ID, item.UserID).
		Select("*").
		Omit("DeletedAt").
		Updates(item)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return common.ErrMissing
	}
	return nil
}

// Delete marks the piece deleted. The row stays, and reads leave it out.
func (r *repository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	res := r.WithContext(ctx).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Delete(&Item{})
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
	query := r.WithContext(ctx).Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID)
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
