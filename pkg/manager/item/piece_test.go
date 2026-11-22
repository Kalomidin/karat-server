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
		Purity:      "18K",
		WeightGrams: 3.2,
		CurrencyID:  currency.DollarID,
		Cost:        50,
	}
}

func storedPiece() *repository.Item {
	created, _ := parseDay("2026-11-21")
	return &repository.Item{
		Name:        "Oval sapphire ring",
		Kind:        "ring",
		Metal:       "gold",
		Purity:      "18K",
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
	if err := applySale(item, "2026-11-22", 150, "2026-11-21"); err != common.ErrFutureSale {
		t.Fatalf("future: %v", err)
	}
	if err := applySale(item, "2026-11-21", 150, "2026-11-21"); err != nil {
		t.Fatal(err)
	}
	if err := applySale(item, "2026-11-21", 160, "2026-11-21"); err != common.ErrAlreadySold {
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
	if dayString(time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)) != "2026-11-21" {
		t.Fatal("date format")
	}
}
