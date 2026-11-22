package domain

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"
)

const maxCents = 10_000_000_000

// Item is one piece in a user's stock book.
//
// A sale is recorded only when both SoldOn and SoldCents are set.
type Item struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	StockNumber string  `json:"stockNumber"`
	Kind        string  `json:"kind"`
	Metal       string  `json:"metal"`
	Purity      string  `json:"purity"`
	WeightGrams float64 `json:"weightGrams"`
	CostCents   int     `json:"costCents"`
	Notes       string  `json:"notes"`
	CreatedOn   string  `json:"createdOn"`
	SoldOn      *string `json:"soldOn"`
	SoldCents   *int    `json:"soldCents"`
}

// Input is the editable detail of a piece, before a sale.
type Input struct {
	Name        string
	StockNumber string
	Kind        string
	Metal       string
	Purity      string
	WeightGrams float64
	CostCents   int
	Notes       string
}

var kinds = map[string]struct{}{
	"ring": {}, "necklace": {}, "bracelet": {}, "earrings": {},
	"pendant": {}, "brooch": {}, "other": {},
}

var metals = map[string]struct{}{
	"gold": {}, "silver": {}, "platinum": {}, "mixed": {},
}

// Normalize trims text, rounds grams to three decimals, and rejects a piece
// the stock book cannot store.
func Normalize(in Input) (Input, error) {
	in.Name = trim(in.Name)
	in.StockNumber = trim(in.StockNumber)
	in.Kind = trim(in.Kind)
	in.Metal = trim(in.Metal)
	in.Purity = trim(in.Purity)
	in.Notes = trim(in.Notes)

	switch {
	case in.Name == "" || tooLong(in.Name, 200):
		return Input{}, ErrInvalidName
	case tooLong(in.StockNumber, 80):
		return Input{}, ErrInvalidStockNumber
	case !known(kinds, in.Kind):
		return Input{}, ErrInvalidKind
	case !known(metals, in.Metal):
		return Input{}, ErrInvalidMetal
	case in.Purity == "" || tooLong(in.Purity, 40):
		return Input{}, ErrInvalidPurity
	case !gramsOK(in.WeightGrams):
		return Input{}, ErrInvalidWeight
	case in.CostCents < 0 || in.CostCents > maxCents:
		return Input{}, ErrInvalidCost
	case tooLong(in.Notes, 4000):
		return Input{}, ErrInvalidNotes
	}
	in.WeightGrams = math.Round(in.WeightGrams*1000) / 1000
	return in, nil
}

// NewItem builds an in-stock piece registered on today.
func NewItem(id, today string, in Input) (Item, error) {
	normalized, err := Normalize(in)
	if err != nil {
		return Item{}, err
	}
	if !ValidDay(today) {
		return Item{}, ErrInvalidDate
	}
	return Item{
		ID:          id,
		Name:        normalized.Name,
		StockNumber: normalized.StockNumber,
		Kind:        normalized.Kind,
		Metal:       normalized.Metal,
		Purity:      normalized.Purity,
		WeightGrams: normalized.WeightGrams,
		CostCents:   normalized.CostCents,
		Notes:       normalized.Notes,
		CreatedOn:   today,
	}, nil
}

// Edit replaces the piece details. A sold piece stays as it was sold.
func Edit(item Item, in Input) (Item, error) {
	if item.hasSale() {
		return Item{}, ErrSoldLocked
	}
	normalized, err := Normalize(in)
	if err != nil {
		return Item{}, err
	}
	item.Name = normalized.Name
	item.StockNumber = normalized.StockNumber
	item.Kind = normalized.Kind
	item.Metal = normalized.Metal
	item.Purity = normalized.Purity
	item.WeightGrams = normalized.WeightGrams
	item.CostCents = normalized.CostCents
	item.Notes = normalized.Notes
	return item, nil
}

// Sell records the price and the calendar day of the sale.
func Sell(item Item, soldOn string, cents int, today string) (Item, error) {
	if item.hasSale() {
		return Item{}, ErrAlreadySold
	}
	if !ValidDay(soldOn) || !ValidDay(today) {
		return Item{}, ErrInvalidDate
	}
	if soldOn > today {
		return Item{}, ErrFutureSale
	}
	if cents < 0 || cents > maxCents {
		return Item{}, ErrInvalidPrice
	}
	soldOnCopy := soldOn
	centsCopy := cents
	item.SoldOn = &soldOnCopy
	item.SoldCents = &centsCopy
	return item, nil
}

// Unsell clears the sale so the piece is in stock again.
func Unsell(item Item) (Item, error) {
	if !item.hasSale() {
		return Item{}, ErrStillInStock
	}
	item.SoldOn = nil
	item.SoldCents = nil
	return item, nil
}

func (item Item) hasSale() bool {
	return item.SoldOn != nil || item.SoldCents != nil
}

// ValidDay reports whether value is a real calendar day, YYYY-MM-DD.
func ValidDay(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

// Clone copies sale fields so a stored piece cannot be changed through a returned value.
func Clone(item Item) Item {
	if item.SoldOn != nil {
		soldOn := *item.SoldOn
		item.SoldOn = &soldOn
	}
	if item.SoldCents != nil {
		cents := *item.SoldCents
		item.SoldCents = &cents
	}
	return item
}

func trim(value string) string {
	return strings.TrimSpace(value)
}

func tooLong(value string, max int) bool {
	return utf8.RuneCountInString(value) > max
}

func known(set map[string]struct{}, value string) bool {
	_, ok := set[value]
	return ok
}

func gramsOK(grams float64) bool {
	if math.IsNaN(grams) || math.IsInf(grams, 0) {
		return false
	}
	rounded := math.Round(grams*1000) / 1000
	return rounded > 0 && rounded <= 100000
}
