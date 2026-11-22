package user_manager

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `json:"-"`
	GoogleID string    `json:"googleId"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	PhotoURL string    `json:"photoUrl,omitempty"`
}
