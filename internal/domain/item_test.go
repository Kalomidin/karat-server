package domain

import "testing"

func sample() Input {
	return Input{
		Name:        " Oval sapphire ring ",
		StockNumber: "A-1",
		Kind:        "ring",
		Metal:       "gold",
		Purity:      "18K",
		WeightGrams: 3.2,
		CostCents:   5000,
		Notes:       "blue",
	}
}

func TestNormalizeTrimsAndRounds(t *testing.T) {
	got, err := Normalize(sample())
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "Oval sapphire ring" {
		t.Fatalf("name %q", got.Name)
	}
	in := sample()
	in.WeightGrams = 3.2346
	got, err = Normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.WeightGrams != 3.235 {
		t.Fatalf("grams %v", got.WeightGrams)
	}
}

func TestNormalizeRejectsUnknownKind(t *testing.T) {
	in := sample()
	in.Kind = "tiara"
	if _, err := Normalize(in); err != ErrInvalidKind {
		t.Fatalf("got %v", err)
	}
}

func TestSellRejectsAFutureDayAndASecondSale(t *testing.T) {
	item, err := NewItem("piece", "2026-11-21", sample())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Sell(item, "2026-11-22", 15000, "2026-11-21"); err != ErrFutureSale {
		t.Fatalf("future: %v", err)
	}
	sold, err := Sell(item, "2026-11-21", 15000, "2026-11-21")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Sell(sold, "2026-11-21", 16000, "2026-11-21"); err != ErrAlreadySold {
		t.Fatalf("second sale: %v", err)
	}
	if _, err := Edit(sold, sample()); err != ErrSoldLocked {
		t.Fatalf("edit: %v", err)
	}
	back, err := Unsell(sold)
	if err != nil {
		t.Fatal(err)
	}
	if back.SoldOn != nil || back.SoldCents != nil {
		t.Fatal("sale was kept")
	}
}
