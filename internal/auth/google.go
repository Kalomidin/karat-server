// Package auth checks the Google ID token a Karat client sends.
package auth

import (
	"context"
	"log"
	"strings"

	"google.golang.org/api/idtoken"
)

// Identity is the Google account inside a verified token.
// Subject is Google's stable account id.
type Identity struct {
	Subject  string
	Email    string
	Name     string
	PhotoURL string
}

// Authenticator turns a bearer token into an account.
type Authenticator interface {
	Authenticate(ctx context.Context, raw string) (Identity, error)
}

// ErrUnauthorized means the token is missing, expired, or not for this server.
var ErrUnauthorized = &unauthorized{}

type unauthorized struct{}

func (unauthorized) Error() string { return "unauthorized" }

// Google checks ID tokens against the configured OAuth client ids.
type Google struct {
	audiences []string
	dev       bool
}

// NewGoogle accepts tokens minted for any of audiences.
// When dev is set, a local token of the form dev:<id>:<email>:<name> is also accepted.
func NewGoogle(audiences []string, dev bool) *Google {
	return &Google{audiences: audiences, dev: dev}
}

// NewDev accepts only local dev tokens. Tests use it so they do not call Google.
func NewDev() *Google {
	return &Google{dev: true}
}

func (g *Google) Authenticate(ctx context.Context, raw string) (Identity, error) {
	if g.dev {
		if identity, ok := devIdentity(raw); ok {
			return identity, nil
		}
	}
	var last error
	for _, audience := range g.audiences {
		payload, err := idtoken.Validate(ctx, raw, audience)
		if err != nil {
			last = err
			continue
		}
		identity, err := identityFromClaims(payload.Subject, payload.Claims)
		if err != nil {
			return Identity{}, err
		}
		return identity, nil
	}
	if last != nil {
		log.Printf("google token rejected: %v", last)
	}
	return Identity{}, ErrUnauthorized
}

func identityFromClaims(subject string, claims map[string]any) (Identity, error) {
	email := claimString(claims, "email")
	if subject == "" || email == "" || !emailVerified(claims["email_verified"]) {
		return Identity{}, ErrUnauthorized
	}
	return Identity{
		Subject:  subject,
		Email:    email,
		Name:     personName(claims),
		PhotoURL: claimString(claims, "picture"),
	}, nil
}

func personName(claims map[string]any) string {
	if name := claimString(claims, "name"); name != "" {
		return name
	}
	given := claimString(claims, "given_name")
	family := claimString(claims, "family_name")
	return strings.TrimSpace(given + " " + family)
}

func claimString(claims map[string]any, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}

func emailVerified(claim any) bool {
	switch value := claim.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	default:
		return false
	}
}

func devIdentity(raw string) (Identity, bool) {
	parts := strings.Split(raw, ":")
	if len(parts) < 3 || parts[0] != "dev" {
		return Identity{}, false
	}
	id := parts[1]
	email := parts[2]
	name := ""
	if len(parts) > 3 {
		name = strings.Join(parts[3:], ":")
	}
	if !devID(id) || !strings.Contains(email, "@") || strings.Contains(email, " ") {
		return Identity{}, false
	}
	return Identity{Subject: id, Email: email, Name: strings.TrimSpace(name)}, true
}

func devID(id string) bool {
	if id == "" || len(id) > 80 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.':
		default:
			return false
		}
	}
	return true
}
