package item_handler

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"karat-server/common"
	manager "karat-server/pkg/manager/item"
)

func readItem(ctx *gin.Context) (itemRequest, error) {
	var body itemRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return itemRequest{}, common.ErrInvalidJSON
	}
	switch {
	case body.Name == nil:
		return itemRequest{}, common.ErrInvalidName
	case body.Kind == nil:
		return itemRequest{}, common.ErrInvalidKind
	case body.Metal == nil:
		return itemRequest{}, common.ErrInvalidMetal
	case body.Purity == nil:
		return itemRequest{}, common.ErrInvalidPurity
	case body.WeightGrams == nil:
		return itemRequest{}, common.ErrInvalidWeight
	case body.CurrencyID == nil:
		return itemRequest{}, common.ErrInvalidCurrency
	case body.Cost == nil:
		return itemRequest{}, common.ErrInvalidCost
	}
	return body, nil
}

func (body itemRequest) input() (manager.Input, error) {
	currencyID, err := uuid.Parse(strings.TrimSpace(*body.CurrencyID))
	if err != nil {
		return manager.Input{}, common.ErrInvalidCurrency
	}
	in := manager.Input{
		Name:        *body.Name,
		Kind:        *body.Kind,
		Metal:       *body.Metal,
		Purity:      *body.Purity,
		WeightGrams: *body.WeightGrams,
		CurrencyID:  currencyID,
		Cost:        *body.Cost,
	}
	if body.Notes != nil {
		in.Notes = *body.Notes
	}
	if body.CreatedOn != nil {
		created, err := time.Parse(time.RFC3339, strings.TrimSpace(*body.CreatedOn))
		if err != nil {
			return manager.Input{}, common.ErrInvalidDate
		}
		in.CreatedOn = created.UTC()
	}
	in.ImageKey = body.ImageKey
	in.Quantity = 1
	if body.Quantity != nil {
		in.Quantity = *body.Quantity
	}
	return in, nil
}
