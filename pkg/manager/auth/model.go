package auth_manager

type Session struct {
	GoogleID string `json:"googleId"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	PhotoURL string `json:"photoUrl,omitempty"`
}
