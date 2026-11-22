package item_manager

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"karat-server/common"
	"karat-server/pkg/manager/item/repository"
)

const maxAmount = 1_000_000_000

func normalize(in Input) (Input, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Kind = strings.TrimSpace(in.Kind)
	in.Metal = strings.TrimSpace(in.Metal)
	in.Purity = strings.TrimSpace(in.Purity)
	in.Notes = strings.TrimSpace(in.Notes)
	switch {
	case in.Name == "" || utf8.RuneCountInString(in.Name) > 200:
		return Input{}, common.ErrInvalidName
	case !knownKind(in.Kind):
		return Input{}, common.ErrInvalidKind
	case !knownMetal(in.Metal):
		return Input{}, common.ErrInvalidMetal
	case in.Purity == "" || utf8.RuneCountInString(in.Purity) > 40:
		return Input{}, common.ErrInvalidPurity
	case in.CurrencyID == uuid.Nil:
		return Input{}, common.ErrInvalidCurrency
	case !gramsOK(in.WeightGrams):
		return Input{}, common.ErrInvalidWeight
	case in.Cost < 0 || in.Cost > maxAmount:
		return Input{}, common.ErrInvalidCost
	case utf8.RuneCountInString(in.Notes) > 4000:
		return Input{}, common.ErrInvalidNotes
	}
	in.WeightGrams = math.Round(in.WeightGrams*1000) / 1000
	return in, nil
}

func applyEdit(item *repository.Item, in Input) error {
	if hasSale(item) {
		return common.ErrSoldLocked
	}
	normalized, err := normalize(in)
	if err != nil {
		return err
	}
	item.Name = normalized.Name
	item.Kind = normalized.Kind
	item.Metal = normalized.Metal
	item.Purity = normalized.Purity
	item.CurrencyID = normalized.CurrencyID
	item.WeightGrams = normalized.WeightGrams
	item.Cost = int64(normalized.Cost)
	item.Notes = normalized.Notes
	if in.ImageKey != nil {
		item.ImageKey = strings.TrimSpace(*in.ImageKey)
	}
	return nil
}

func applySale(item *repository.Item, soldOn string, amount int, today string) error {
	if hasSale(item) {
		return common.ErrAlreadySold
	}
	sold, err := parseDay(soldOn)
	if err != nil || !validDay(today) {
		return common.ErrInvalidDate
	}
	if soldOn > today {
		return common.ErrFutureSale
	}
	if amount < 0 || amount > maxAmount {
		return common.ErrInvalidPrice
	}
	soldAmount := int64(amount)
	item.SoldOn = &sold
	item.SoldAmount = &soldAmount
	return nil
}

func applyReturn(item *repository.Item) error {
	if !hasSale(item) {
		return common.ErrStillInStock
	}
	item.SoldOn = nil
	item.SoldAmount = nil
	return nil
}

func hasSale(item *repository.Item) bool {
	return item.SoldOn != nil || item.SoldAmount != nil
}

func validDay(value string) bool {
	parsed, err := time.Parse("2006-01-02", value)
	return err == nil && parsed.Format("2006-01-02") == value
}

func parseDay(value string) (time.Time, error) {
	if !validDay(value) {
		return time.Time{}, common.ErrInvalidDate
	}
	return time.Parse("2006-01-02", value)
}

func dayString(value time.Time) string {
	year, month, day := value.Date()
	return fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}

func knownKind(kind string) bool {
	switch kind {
	case "ring", "necklace", "bracelet", "earrings", "pendant", "brooch", "other":
		return true
	default:
		return false
	}
}

func knownMetal(metal string) bool {
	switch metal {
	case "gold", "silver", "platinum", "mixed":
		return true
	default:
		return false
	}
}

func gramsOK(grams float64) bool {
	if math.IsNaN(grams) || math.IsInf(grams, 0) {
		return false
	}
	rounded := math.Round(grams*1000) / 1000
	return rounded > 0 && rounded <= 100000
}

func toItem(item *repository.Item) *Item {
	out := &Item{
		ID:          item.ID.String(),
		Name:        item.Name,
		Kind:        item.Kind,
		Metal:       item.Metal,
		Purity:      item.Purity,
		WeightGrams: item.WeightGrams,
		CurrencyID:  item.CurrencyID.String(),
		Cost:        int(item.Cost),
		Notes:       item.Notes,
		ImageKey:    item.ImageKey,
		CreatedOn:   dayString(item.CreatedOn),
	}
	if item.SoldOn != nil {
		soldOn := dayString(*item.SoldOn)
		out.SoldOn = &soldOn
	}
	if item.SoldAmount != nil {
		soldAmount := int(*item.SoldAmount)
		out.SoldAmount = &soldAmount
	}
	return out
}
