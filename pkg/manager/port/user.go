package port

import (
	"context"

	"github.com/google/uuid"
)

type GoogleUser struct {
	GoogleID string
	Email    string
	Name     string
	PhotoURL string
}

type User struct {
	ID       uuid.UUID
	GoogleID string
	Email    string
	Name     string
	PhotoURL string
}

type UserPort interface {
	UpsertGoogleUser(ctx context.Context, user GoogleUser) (*User, error)
}
