package domain

// User is the Google account that owns a stock book.
// GoogleID is the stable subject from the Google ID token.
type User struct {
	GoogleID string `json:"googleId"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	PhotoURL string `json:"photoUrl,omitempty"`
}
