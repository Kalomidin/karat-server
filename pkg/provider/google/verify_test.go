package google

import "testing"

func TestGoogleClaimsKeepTheAccount(t *testing.T) {
	got, err := userFromClaims("109820475938", map[string]interface{}{
		"email":          "ada@atelier.test",
		"email_verified": true,
		"given_name":     "Ada",
		"family_name":    "Lovelace",
		"picture":        "https://example.test/ada.jpg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.GoogleID != "109820475938" || got.Email != "ada@atelier.test" || got.Name != "Ada Lovelace" {
		t.Fatalf("%+v", got)
	}
	if got.PhotoURL != "https://example.test/ada.jpg" {
		t.Fatalf("photo %q", got.PhotoURL)
	}
}

func TestDevLogin(t *testing.T) {
	client := NewClient(Config{DevLogin: true})
	got, err := client.VerifyIDToken(t.Context(), "dev:ada:ada@atelier.test:Ada Lovelace")
	if err != nil {
		t.Fatal(err)
	}
	if got.GoogleID != "ada" || got.Email != "ada@atelier.test" || got.Name != "Ada Lovelace" {
		t.Fatalf("%+v", got)
	}
}
