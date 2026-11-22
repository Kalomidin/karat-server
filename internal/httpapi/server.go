// Package httpapi is the HTTP API for the Karat server.
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"karat-server/internal/auth"
	"karat-server/internal/domain"
	"karat-server/internal/store"
)

type contextKey struct{}

// Option changes server behavior. Tests use it to freeze the clock.
type Option func(*handler)

// WithClock sets the clock used for "today" when a sale is recorded.
func WithClock(now func() time.Time) Option {
	return func(h *handler) { h.now = now }
}

type handler struct {
	store store.Store
	auth  auth.Authenticator
	now   func() time.Time
	loc   *time.Location
}

// New serves the stock book. A Google ID token identifies the caller.
func New(books store.Store, tokens auth.Authenticator, loc *time.Location, opts ...Option) http.Handler {
	if loc == nil {
		loc = time.UTC
	}
	h := &handler{store: books, auth: tokens, now: time.Now, loc: loc}
	for _, opt := range opts {
		opt(h)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.health)
	mux.HandleFunc("POST /v1/session", h.withUser(h.session))
	mux.HandleFunc("GET /v1/me", h.withUser(h.session))
	mux.HandleFunc("GET /v1/items", h.withUser(h.listItems))
	mux.HandleFunc("POST /v1/items", h.withUser(h.createItem))
	mux.HandleFunc("GET /v1/items/{id}", h.withUser(h.getItem))
	mux.HandleFunc("PUT /v1/items/{id}", h.withUser(h.updateItem))
	mux.HandleFunc("DELETE /v1/items/{id}", h.withUser(h.deleteItem))
	mux.HandleFunc("POST /v1/items/{id}/sale", h.withUser(h.sellItem))
	mux.HandleFunc("POST /v1/items/{id}/return", h.withUser(h.returnItem))
	return h.log(mux)
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]bool{"ok": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *handler) session(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, userFrom(r.Context()))
}

func (h *handler) listItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.store.ListItems(r.Context(), userFrom(r.Context()).GoogleID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *handler) createItem(w http.ResponseWriter, r *http.Request) {
	input, err := readItem(w, r)
	if err != nil {
		writeError(w, err)
		return
	}
	item, err := h.store.CreateItem(r.Context(), userFrom(r.Context()).GoogleID, input, h.today())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *handler) getItem(w http.ResponseWriter, r *http.Request) {
	item, err := h.store.GetItem(r.Context(), userFrom(r.Context()).GoogleID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *handler) updateItem(w http.ResponseWriter, r *http.Request) {
	input, err := readItem(w, r)
	if err != nil {
		writeError(w, err)
		return
	}
	item, err := h.store.UpdateItem(r.Context(), userFrom(r.Context()).GoogleID, r.PathValue("id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	err := h.store.DeleteItem(r.Context(), userFrom(r.Context()).GoogleID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) sellItem(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SoldOn    *string `json:"soldOn"`
		SoldCents *int    `json:"soldCents"`
	}
	if err := decode(w, r, &body); err != nil {
		writeError(w, err)
		return
	}
	if body.SoldOn == nil {
		writeError(w, domain.ErrInvalidDate)
		return
	}
	if body.SoldCents == nil {
		writeError(w, domain.ErrInvalidPrice)
		return
	}
	item, err := h.store.MarkSold(
		r.Context(),
		userFrom(r.Context()).GoogleID,
		r.PathValue("id"),
		*body.SoldOn,
		*body.SoldCents,
		h.today(),
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *handler) returnItem(w http.ResponseWriter, r *http.Request) {
	item, err := h.store.ReturnToStock(r.Context(), userFrom(r.Context()).GoogleID, r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *handler) withUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearer(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, auth.ErrUnauthorized)
			return
		}
		identity, err := h.auth.Authenticate(r.Context(), token)
		if err != nil {
			writeError(w, auth.ErrUnauthorized)
			return
		}
		user, err := h.store.UpsertUser(r.Context(), domain.User{
			GoogleID: identity.Subject,
			Email:    identity.Email,
			Name:     identity.Name,
			PhotoURL: identity.PhotoURL,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), contextKey{}, user)
		next(w, r.WithContext(ctx))
	}
}

func (h *handler) today() string {
	current := h.now().In(h.loc)
	return fmt.Sprintf("%04d-%02d-%02d", current.Year(), int(current.Month()), current.Day())
}

func userFrom(ctx context.Context) domain.User {
	user, _ := ctx.Value(contextKey{}).(domain.User)
	return user
}

type itemBody struct {
	Name        *string  `json:"name"`
	StockNumber *string  `json:"stockNumber"`
	Kind        *string  `json:"kind"`
	Metal       *string  `json:"metal"`
	Purity      *string  `json:"purity"`
	WeightGrams *float64 `json:"weightGrams"`
	CostCents   *int     `json:"costCents"`
	Notes       *string  `json:"notes"`
}

func readItem(w http.ResponseWriter, r *http.Request) (domain.Input, error) {
	var body itemBody
	if err := decode(w, r, &body); err != nil {
		return domain.Input{}, err
	}
	switch {
	case body.Name == nil:
		return domain.Input{}, domain.ErrInvalidName
	case body.Kind == nil:
		return domain.Input{}, domain.ErrInvalidKind
	case body.Metal == nil:
		return domain.Input{}, domain.ErrInvalidMetal
	case body.Purity == nil:
		return domain.Input{}, domain.ErrInvalidPurity
	case body.WeightGrams == nil:
		return domain.Input{}, domain.ErrInvalidWeight
	case body.CostCents == nil:
		return domain.Input{}, domain.ErrInvalidCost
	}
	input := domain.Input{
		Name:        *body.Name,
		Kind:        *body.Kind,
		Metal:       *body.Metal,
		Purity:      *body.Purity,
		WeightGrams: *body.WeightGrams,
		CostCents:   *body.CostCents,
	}
	if body.StockNumber != nil {
		input.StockNumber = *body.StockNumber
	}
	if body.Notes != nil {
		input.Notes = *body.Notes
	}
	return input, nil
}

func decode(w http.ResponseWriter, r *http.Request, dest any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return domain.ErrInvalidJSON
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return domain.ErrInvalidJSON
	}
	return nil
}

func bearer(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", false
	}
	return token, true
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrUnauthorized) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	var client *domain.Error
	if errors.As(err, &client) {
		writeJSON(w, statusOf(client.Code), map[string]string{"error": client.Code})
		return
	}
	log.Printf("request failed: %v", err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal"})
}

func statusOf(code string) int {
	switch code {
	case domain.ErrMissing.Code:
		return http.StatusNotFound
	case domain.ErrAlreadySold.Code, domain.ErrSoldLocked.Code, domain.ErrStillInStock.Code:
		return http.StatusConflict
	default:
		return http.StatusBadRequest
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	_ = encoder.Encode(body)
}

func (h *handler) log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, wrapped.status, time.Since(start).Round(time.Millisecond))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
