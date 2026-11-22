package user_manager

import (
	"context"

	"karat-server/pkg/manager/port"
	"karat-server/pkg/manager/user/repository"
)

type UserManager interface {
	UpsertGoogleUser(ctx context.Context, user port.GoogleUser) (*port.User, error)
	Get(ctx context.Context, user port.User) (*User, error)
}

type userManager struct {
	repository repository.Repository
}

func NewUserManager(repository repository.Repository) UserManager {
	return &userManager{repository: repository}
}

func NewUserPort(repository repository.Repository) port.UserPort {
	return &userManager{repository: repository}
}

func (m *userManager) UpsertGoogleUser(ctx context.Context, user port.GoogleUser) (*port.User, error) {
	stored, err := m.repository.UpsertByGoogleID(ctx, &repository.User{
		GoogleID: user.GoogleID,
		Email:    user.Email,
		Name:     user.Name,
		PhotoURL: user.PhotoURL,
	})
	if err != nil {
		return nil, err
	}
	return toPort(stored), nil
}

func (m *userManager) Get(_ context.Context, user port.User) (*User, error) {
	return &User{
		ID:       user.ID,
		GoogleID: user.GoogleID,
		Email:    user.Email,
		Name:     user.Name,
		PhotoURL: user.PhotoURL,
	}, nil
}

func toPort(user *repository.User) *port.User {
	return &port.User{
		ID:       user.ID,
		GoogleID: user.GoogleID,
		Email:    user.Email,
		Name:     user.Name,
		PhotoURL: user.PhotoURL,
	}
}
