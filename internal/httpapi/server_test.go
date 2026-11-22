package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"karat-server/internal/auth"
	"karat-server/internal/store"
)

func TestLoginStoresGoogleIdentity(t *testing.T) {
	api := testAPI(t)
	response := call(api, http.MethodPost, "/v1/session", "dev:109820475938:ada@atelier.test:Ada", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status %d body %s", response.Code, response.Body.String())
	}
	var user struct {
		GoogleID string `json:"googleId"`
		Email    string `json:"email"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	if user.GoogleID != "109820475938" || user.Email != "ada@atelier.test" || user.Name != "Ada" {
		t.Fatalf("%+v", user)
	}

	again := call(api, http.MethodGet, "/v1/me", "dev:109820475938:ada@atelier.test:Ada Atelier", "")
	if err := json.Unmarshal(again.Body.Bytes(), &user); err != nil {
		t.Fatal(err)
	}
	if user.Name != "Ada Atelier" || user.GoogleID != "109820475938" {
		t.Fatalf("profile was not updated: %+v", user)
	}
}

func TestItemsStayWithTheGoogleAccount(t *testing.T) {
	api := testAPI(t)
	ada := "dev:ada:ada@atelier.test:Ada"
	created := call(api, http.MethodPost, "/v1/items", ada, `{
		"name": "Oval sapphire ring",
		"kind": "ring",
		"metal": "gold",
		"purity": "18K",
		"weightGrams": 3.2,
		"costCents": 5000,
		"notes": ""
	}`)
	if created.Code != http.StatusCreated {
		t.Fatalf("create %d %s", created.Code, created.Body.String())
	}
	var item struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}

	other := call(api, http.MethodGet, "/v1/items/"+item.ID, "dev:bea:bea@atelier.test:Bea", "")
	if other.Code != http.StatusNotFound {
		t.Fatalf("other user status %d", other.Code)
	}

	future := call(api, http.MethodPost, "/v1/items/"+item.ID+"/sale", ada, `{"soldOn":"2026-11-22","soldCents":15000}`)
	if future.Code != http.StatusBadRequest {
		t.Fatalf("future sale %d %s", future.Code, future.Body.String())
	}
	sold := call(api, http.MethodPost, "/v1/items/"+item.ID+"/sale", ada, `{"soldOn":"2026-11-21","soldCents":15000}`)
	if sold.Code != http.StatusOK {
		t.Fatalf("sale %d %s", sold.Code, sold.Body.String())
	}
	missing := call(api, http.MethodPost, "/v1/items", "", `{}`)
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("auth %d", missing.Code)
	}
}

func testAPI(t *testing.T) http.Handler {
	t.Helper()
	return New(store.NewMemory(), auth.NewDev(), time.UTC, WithClock(func() time.Time {
		return time.Date(2026, 11, 21, 15, 0, 0, 0, time.UTC)
	}))
}

func call(api http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	api.ServeHTTP(recorder, request)
	return recorder
}
