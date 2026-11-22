// Package store keeps users and their jewellery pieces.
package store

import (
	"context"
	"crypto/rand"
	"fmt"

	"karat-server/internal/domain"
)

// Store is the stock book for every signed-in user.
type Store interface {
	Ping(ctx context.Context) error
	UpsertUser(ctx context.Context, user domain.User) (domain.User, error)
	ListItems(ctx context.Context, userID string) ([]domain.Item, error)
	CreateItem(ctx context.Context, userID string, in domain.Input, today string) (domain.Item, error)
	GetItem(ctx context.Context, userID, id string) (domain.Item, error)
	UpdateItem(ctx context.Context, userID, id string, in domain.Input) (domain.Item, error)
	MarkSold(ctx context.Context, userID, id, soldOn string, soldCents int, today string) (domain.Item, error)
	ReturnToStock(ctx context.Context, userID, id string) (domain.Item, error)
	DeleteItem(ctx context.Context, userID, id string) error
}

func newID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
