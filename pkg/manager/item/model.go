package item_manager

import (
	"time"

	"github.com/google/uuid"
)

type Input struct {
	Name        string
	Kind        string
	Metal       string
	Purity      string
	WeightGrams float64
	CurrencyID  uuid.UUID
	Cost        int
	Notes       string
	ImageKey    *string
	CreatedOn   time.Time
	Quantity    int
	GroupID     uuid.UUID
}

type Item struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Kind        string  `json:"kind"`
	Metal       string  `json:"metal"`
	Purity      string  `json:"purity"`
	WeightGrams float64 `json:"weightGrams"`
	CurrencyID  string  `json:"currencyId"`
	Cost        int     `json:"cost"`
	Notes       string  `json:"notes"`
	ImageKey    string  `json:"imageKey,omitempty"`
	ImageURL    string  `json:"imageUrl,omitempty"`
	CreatedOn   string  `json:"createdOn"`
	GroupID     string  `json:"groupId,omitempty"`
	SoldOn      *string `json:"soldOn"`
	SoldAmount  *int    `json:"soldAmount"`
}

type Sale struct {
	SoldOn     string
	SoldAmount int
}
