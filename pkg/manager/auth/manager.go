package auth_manager

import (
	"context"

	"karat-server/common"
	"karat-server/pkg/manager/port"
	"karat-server/pkg/provider"
)

type AuthManager interface {
	Session(ctx context.Context, idToken string) (*Session, error)
}

type authManager struct {
	users    port.UserPort
	provider provider.Client
}

func NewAuthManager(users port.UserPort, provider provider.Client) AuthManager {
	return &authManager{users: users, provider: provider}
}

func (m *authManager) Session(ctx context.Context, idToken string) (*Session, error) {
	details, err := m.provider.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, common.ErrUnauthorized
	}
	user, err := m.users.UpsertGoogleUser(ctx, port.GoogleUser{
		GoogleID: details.GoogleID,
		Email:    details.Email,
		Name:     details.Name,
		PhotoURL: details.PhotoURL,
	})
	if err != nil {
		return nil, err
	}
	return &Session{
		GoogleID: user.GoogleID,
		Email:    user.Email,
		Name:     user.Name,
		PhotoURL: user.PhotoURL,
	}, nil
}
