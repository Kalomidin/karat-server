package common

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	GoogleID string
	Email    string
	Name     string
	PhotoURL string
}

type userKey struct{}

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

func CurrentUser(ctx context.Context) (User, error) {
	user, ok := ctx.Value(userKey{}).(User)
	if !ok {
		return User{}, ErrUnauthorized
	}
	return user, nil
}
