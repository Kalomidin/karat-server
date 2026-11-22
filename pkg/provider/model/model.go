package model

// UserDetails is the Google account taken from a verified login token.
type UserDetails struct {
	GoogleID string
	Email    string
	Name     string
	PhotoURL string
}
