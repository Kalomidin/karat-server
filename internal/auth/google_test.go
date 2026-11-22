package auth

import "testing"

func TestGoogleClaimsKeepTheAccount(t *testing.T) {
	got, err := identityFromClaims("109820475938", map[string]any{
		"email":          "ada@atelier.test",
		"email_verified": true,
		"given_name":     "Ada",
		"family_name":    "Lovelace",
		"picture":        "https://example.test/ada.jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != "109820475938" || got.Email != "ada@atelier.test" || got.Name != "Ada Lovelace" {
		t.Fatalf("%+v", got)
	}
	if got.PhotoURL != "https://example.test/ada.jpg" {
		t.Fatalf("photo %q", got.PhotoURL)
	}

	named, err := identityFromClaims("109820475938", map[string]any{
		"email":          "ada@atelier.test",
		"email_verified": "true",
		"name":           "Ada",
		"given_name":     "Augusta",
	})
	if err != nil {
		t.Fatal(err)
	}
	if named.Name != "Ada" {
		t.Fatalf("name %q", named.Name)
	}
}

func TestDevToken(t *testing.T) {
	got, err := NewDev().Authenticate(t.Context(), "dev:ada:ada@atelier.test:Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	if got.Subject != "ada" || got.Email != "ada@atelier.test" || got.Name != "Ada Lovelace" {
		t.Fatalf("%+v", got)
	}
	if _, err := NewDev().Authenticate(t.Context(), "not-a-token"); err != ErrUnauthorized {
		t.Fatalf("got %v", err)
	}
}
