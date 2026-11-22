package item_manager

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"karat-server/common"
	"karat-server/pkg/manager/item/repository"
)

type ItemManager interface {
	List(ctx context.Context, userID uuid.UUID) ([]Item, error)
	Create(ctx context.Context, userID uuid.UUID, in Input) ([]Item, error)
	Get(ctx context.Context, userID, id uuid.UUID) (*Item, error)
	Update(ctx context.Context, userID, id uuid.UUID, in Input) (*Item, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Sell(ctx context.Context, userID, id uuid.UUID, sale Sale) (*Item, error)
	ReturnToStock(ctx context.Context, userID, id uuid.UUID) (*Item, error)
}

type Currencies interface {
	Require(ctx context.Context, id uuid.UUID) error
}

type itemManager struct {
	repository repository.Repository
	currencies Currencies
	now        func() time.Time
	readURL    func(context.Context, string) (string, error)
}

func NewItemManager(repository repository.Repository, currencies Currencies, readURL func(context.Context, string) (string, error)) ItemManager {
	if readURL == nil {
		readURL = func(context.Context, string) (string, error) { return "", nil }
	}
	return &itemManager{repository: repository, currencies: currencies, now: time.Now, readURL: readURL}
}

func (m *itemManager) List(ctx context.Context, userID uuid.UUID) ([]Item, error) {
	stored, err := m.repository.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(stored))
	for i := range stored {
		presented, err := m.present(ctx, &stored[i])
		if err != nil {
			return nil, err
		}
		items = append(items, *presented)
	}
	return items, nil
}

func (m *itemManager) Create(ctx context.Context, userID uuid.UUID, in Input) ([]Item, error) {
	if err := validQuantity(in.Quantity); err != nil {
		return nil, err
	}
	normalized, err := normalize(in)
	if err != nil {
		return nil, err
	}
	if err := m.currencies.Require(ctx, normalized.CurrencyID); err != nil {
		return nil, err
	}
	if in.CreatedOn.IsZero() || in.CreatedOn.After(m.now().UTC()) {
		return nil, common.ErrInvalidDate
	}
	groupID := in.GroupID
	if groupID == uuid.Nil {
		groupID = uuid.New()
	}
	var stored []repository.Item
	err = m.repository.WithinTransaction(ctx, func(repo repository.Repository) error {
		stored = make([]repository.Item, 0, in.Quantity)
		for range in.Quantity {
			batch := groupID
			item := repository.Item{
				UserID:      userID,
				Name:        normalized.Name,
				Kind:        normalized.Kind,
				Metal:       normalized.Metal,
				Purity:      normalized.Purity,
				WeightGrams: normalized.WeightGrams,
				CurrencyID:  normalized.CurrencyID,
				Cost:        int64(normalized.Cost),
				Notes:       normalized.Notes,
				GroupID:     &batch,
				CreatedOn:   in.CreatedOn.UTC(),
			}
			if normalized.ImageKey != nil {
				item.ImageKey = strings.TrimSpace(*normalized.ImageKey)
			}
			if err := repo.Create(ctx, &item); err != nil {
				return err
			}
			stored = append(stored, item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	url, err := m.readURL(ctx, stored[0].ImageKey)
	if err != nil {
		return nil, err
	}
	out := make([]Item, 0, len(stored))
	for i := range stored {
		item := toItem(&stored[i])
		item.ImageURL = url
		out = append(out, *item)
	}
	return out, nil
}

func (m *itemManager) Get(ctx context.Context, userID, id uuid.UUID) (*Item, error) {
	item, err := m.repository.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return m.present(ctx, item)
}

func (m *itemManager) Update(ctx context.Context, userID, id uuid.UUID, in Input) (*Item, error) {
	normalized, err := normalize(in)
	if err != nil {
		return nil, err
	}
	if err := m.currencies.Require(ctx, normalized.CurrencyID); err != nil {
		return nil, err
	}
	return m.change(ctx, userID, id, func(item *repository.Item) error {
		return applyEdit(item, normalized)
	})
}

func (m *itemManager) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return m.repository.Delete(ctx, userID, id)
}

func (m *itemManager) Sell(ctx context.Context, userID, id uuid.UUID, sale Sale) (*Item, error) {
	return m.change(ctx, userID, id, func(item *repository.Item) error {
		return applySale(item, sale.SoldOn, sale.SoldAmount, m.now())
	})
}

func (m *itemManager) ReturnToStock(ctx context.Context, userID, id uuid.UUID) (*Item, error) {
	return m.change(ctx, userID, id, func(item *repository.Item) error {
		return applyReturn(item)
	})
}

func (m *itemManager) change(ctx context.Context, userID, id uuid.UUID, apply func(*repository.Item) error) (*Item, error) {
	var updated *repository.Item
	err := m.repository.WithinTransaction(ctx, func(repo repository.Repository) error {
		item, err := repo.GetForUpdate(ctx, userID, id)
		if err != nil {
			return err
		}
		if err := apply(item); err != nil {
			return err
		}
		if err := repo.Save(ctx, item); err != nil {
			return err
		}
		updated = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m.present(ctx, updated)
}

func (m *itemManager) present(ctx context.Context, item *repository.Item) (*Item, error) {
	out := toItem(item)
	url, err := m.readURL(ctx, item.ImageKey)
	if err != nil {
		return nil, err
	}
	out.ImageURL = url
	return out, nil
}
