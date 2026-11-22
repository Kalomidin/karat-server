package google

import (
	"context"
	"log"
	"strings"

	"google.golang.org/api/idtoken"

	"karat-server/pkg/provider/model"
)

func (c *Client) VerifyIDToken(ctx context.Context, raw string) (*model.UserDetails, error) {
	if c.cfg.DevLogin {
		if user, ok := devUser(raw); ok {
			return user, nil
		}
	}
	var last error
	for _, audience := range c.cfg.AudienceList() {
		payload, err := idtoken.Validate(ctx, raw, audience)
		if err != nil {
			last = err
			continue
		}
		user, err := userFromClaims(payload.Subject, payload.Claims)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
	if last != nil {
		log.Printf("google token rejected: %v", last)
	}
	return nil, errUnauthorized
}

func userFromClaims(subject string, claims map[string]interface{}) (*model.UserDetails, error) {
	email := claimString(claims, "email")
	if subject == "" || email == "" || !emailVerified(claims["email_verified"]) {
		return nil, errUnauthorized
	}
	return &model.UserDetails{
		GoogleID: subject,
		Email:    email,
		Name:     personName(claims),
		PhotoURL: claimString(claims, "picture"),
	}, nil
}

func personName(claims map[string]interface{}) string {
	if name := claimString(claims, "name"); name != "" {
		return name
	}
	return strings.TrimSpace(claimString(claims, "given_name") + " " + claimString(claims, "family_name"))
}

func claimString(claims map[string]interface{}, key string) string {
	value, _ := claims[key].(string)
	return strings.TrimSpace(value)
}

func emailVerified(claim interface{}) bool {
	switch value := claim.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(value, "true")
	default:
		return false
	}
}

func devUser(raw string) (*model.UserDetails, bool) {
	parts := strings.Split(raw, ":")
	if len(parts) < 3 || parts[0] != "dev" || !devID(parts[1]) {
		return nil, false
	}
	email := parts[2]
	if !strings.Contains(email, "@") || strings.Contains(email, " ") {
		return nil, false
	}
	name := ""
	if len(parts) > 3 {
		name = strings.TrimSpace(strings.Join(parts[3:], ":"))
	}
	return &model.UserDetails{GoogleID: parts[1], Email: email, Name: name}, true
}

func devID(id string) bool {
	if id == "" || len(id) > 80 {
		return false
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '-', r == '.':
		default:
			return false
		}
	}
	return true
}
