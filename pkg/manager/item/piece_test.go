package item_manager

import (
	"testing"
	"time"

	"karat-server/common"
	"karat-server/pkg/currency"
	"karat-server/pkg/manager/item/repository"
)

func sample() Input {
	return Input{
		Name:        " Oval sapphire ring ",
		Kind:        "ring",
		Metal:       "gold",
		Purity:      "750",
		WeightGrams: 3.2,
		CurrencyID:  currency.DollarID,
		Cost:        50,
	}
}

func storedPiece() *repository.Item {
	created, _ := parseTimestamp("2026-11-21T00:00:00Z")
	return &repository.Item{
		Name:        "Oval sapphire ring",
		Kind:        "ring",
		Metal:       "gold",
		Purity:      "750",
		WeightGrams: 3.2,
		CurrencyID:  currency.DollarID,
		Cost:        50,
		CreatedOn:   created,
	}
}

func TestNormalizeTrimsWeight(t *testing.T) {
	got, err := normalize(sample())
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Oval sapphire ring" {
		t.Fatalf("name %q", got.Name)
	}
	in := sample()
	in.WeightGrams = 3.2346
	got, err = normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.WeightGrams != 3.235 {
		t.Fatalf("grams %v", got.WeightGrams)
	}
}

func TestSaleRules(t *testing.T) {
	item := storedPiece()
	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	if err := applySale(item, "2026-11-22T00:00:00Z", 150, now); err != common.ErrFutureSale {
		t.Fatalf("future: %v", err)
	}
	if err := applySale(item, "2026-11-21T08:00:00Z", 150, now); err != nil {
		t.Fatal(err)
	}
	if err := applySale(item, "2026-11-21T08:00:00Z", 160, now); err != common.ErrAlreadySold {
		t.Fatalf("second: %v", err)
	}
	if err := applyEdit(item, sample()); err != common.ErrSoldLocked {
		t.Fatalf("edit: %v", err)
	}
	if err := applyReturn(item); err != nil {
		t.Fatal(err)
	}
	if item.SoldOn != nil || item.SoldAmount != nil {
		t.Fatal("sale was kept")
	}
	kept := storedPiece()
	kept.ImageKey = "local/items/ada/old.jpg"
	if err := applyEdit(kept, sample()); err != nil {
		t.Fatal(err)
	}
	if kept.ImageKey != "local/items/ada/old.jpg" {
		t.Fatal("image key was cleared")
	}
	next := "local/items/ada/new.jpg"
	edited := sample()
	edited.ImageKey = &next
	if err := applyEdit(kept, edited); err != nil {
		t.Fatal(err)
	}
	if kept.ImageKey != next {
		t.Fatalf("image key %q", kept.ImageKey)
	}
	if timestampString(time.Date(2026, 11, 21, 8, 0, 0, 0, time.FixedZone("shop", 5*60*60))) != "2026-11-21T03:00:00Z" {
		t.Fatal("utc format")
	}
}
