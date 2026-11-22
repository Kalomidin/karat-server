package repository

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"karat-server/common"
	"karat-server/pkg/currency"
)

func TestDeletedPieceStaysOutOfReads(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec(`CREATE TABLE items (
		id text PRIMARY KEY,
		user_id text NOT NULL,
		name text NOT NULL,
		kind text NOT NULL,
		metal text NOT NULL,
		purity text NOT NULL,
		weight_grams real NOT NULL,
		currency_id text NOT NULL,
		cost integer NOT NULL,
		notes text NOT NULL DEFAULT '',
		image_key text NOT NULL DEFAULT '',
		group_id text,
		created_on datetime NOT NULL,
		sold_on datetime,
		sold_amount integer,
		deleted_at datetime,
		created_at datetime,
		updated_at datetime
	)`).Error; err != nil {
		t.Fatal(err)
	}
	repo := NewRepository(db)
	ctx := context.Background()
	user := uuid.New()
	kept := newPiece(user, "Kept ring")
	gone := newPiece(user, "Plain band")
	if err := repo.Create(ctx, kept); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, gone); err != nil {
		t.Fatal(err)
	}

	if err := repo.Delete(ctx, user, gone.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, user, gone.ID); err != common.ErrMissing {
		t.Fatalf("second delete: %v", err)
	}

	listed, err := repo.List(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != kept.ID {
		t.Fatalf("list: %+v", listed)
	}
	if _, err := repo.Get(ctx, user, gone.ID); err != common.ErrMissing {
		t.Fatalf("get: %v", err)
	}
	if _, err := repo.GetForUpdate(ctx, user, gone.ID); err != common.ErrMissing {
		t.Fatalf("lock: %v", err)
	}

	gone.Name = "Changed"
	if err := repo.Save(ctx, gone); err != common.ErrMissing {
		t.Fatalf("save deleted: %v", err)
	}

	var stored Item
	if err := db.Unscoped().First(&stored, "id = ?", gone.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.DeletedAt.Valid {
		t.Fatal("deleted_at was not set")
	}
	if stored.Name != "Plain band" {
		t.Fatalf("name changed to %q", stored.Name)
	}

	sold := time.Date(2026, 11, 21, 8, 0, 0, 0, time.UTC)
	amount := int64(80)
	kept.Notes = "resized"
	kept.SoldOn = &sold
	kept.SoldAmount = &amount
	if err := repo.Save(ctx, kept); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.Get(ctx, user, kept.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Notes != "resized" || loaded.SoldOn == nil || loaded.SoldAmount == nil {
		t.Fatalf("sale was not stored: %+v", loaded)
	}
	if loaded.DeletedAt.Valid {
		t.Fatal("a save marked the piece deleted")
	}
	loaded.SoldOn = nil
	loaded.SoldAmount = nil
	if err := repo.Save(ctx, loaded); err != nil {
		t.Fatal(err)
	}
	returned, err := repo.Get(ctx, user, kept.ID)
	if err != nil {
		t.Fatal(err)
	}
	if returned.SoldOn != nil || returned.SoldAmount != nil {
		t.Fatal("return left the sale in place")
	}
}

func newPiece(user uuid.UUID, name string) *Item {
	return &Item{
		ID:          uuid.New(),
		UserID:      user,
		Name:        name,
		Kind:        "ring",
		Metal:       "gold",
		Purity:      "585",
		WeightGrams: 2,
		CurrencyID:  currency.DollarID,
		Cost:        40,
		CreatedOn:   time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC),
	}
}
