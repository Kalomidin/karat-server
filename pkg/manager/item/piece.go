package item_manager

import (
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
	case in.Metal == "" || utf8.RuneCountInString(in.Metal) > 40:
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

func applySale(item *repository.Item, soldOn string, amount int, now time.Time) error {
	if hasSale(item) {
		return common.ErrAlreadySold
	}
	sold, err := parseTimestamp(soldOn)
	if err != nil {
		return common.ErrInvalidDate
	}
	if sold.After(now.UTC()) {
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

func parseTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, common.ErrInvalidDate
	}
	return parsed.UTC(), nil
}

func timestampString(value time.Time) string {
	return value.UTC().Format(time.RFC3339)
}

func knownKind(kind string) bool {
	switch kind {
	case "ring", "necklace", "bracelet", "earrings", "pendant", "brooch", "other":
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
		CreatedOn:   timestampString(item.CreatedOn),
	}
	if item.SoldOn != nil {
		soldOn := timestampString(*item.SoldOn)
		out.SoldOn = &soldOn
	}
	if item.SoldAmount != nil {
		soldAmount := int(*item.SoldAmount)
		out.SoldAmount = &soldAmount
	}
	return out
}
