package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"karat-server/internal/domain"
)

const itemColumns = `
	id, name, stock_number, kind, metal, purity, weight_grams, cost_cents,
	notes, created_on, sold_on, sold_cents`

// Postgres stores users and pieces in PostgreSQL.
type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) UpsertUser(ctx context.Context, user domain.User) (domain.User, error) {
	const query = `
		INSERT INTO users (google_id, email, name, photo_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_id) DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			photo_url = EXCLUDED.photo_url,
			updated_at = now()
		RETURNING google_id, email, name, photo_url`
	var stored domain.User
	err := p.pool.QueryRow(ctx, query, user.GoogleID, user.Email, user.Name, user.PhotoURL).
		Scan(&stored.GoogleID, &stored.Email, &stored.Name, &stored.PhotoURL)
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert user: %w", err)
	}
	return stored, nil
}

func (p *Postgres) ListItems(ctx context.Context, userID string) ([]domain.Item, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT `+itemColumns+`
		FROM items
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := []domain.Item{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	return items, nil
}

func (p *Postgres) CreateItem(ctx context.Context, userID string, in domain.Input, today string) (domain.Item, error) {
	id, err := newID()
	if err != nil {
		return domain.Item{}, err
	}
	item, err := domain.NewItem(id, today, in)
	if err != nil {
		return domain.Item{}, err
	}
	row := p.pool.QueryRow(ctx, `
		INSERT INTO items (
			id, user_id, name, stock_number, kind, metal, purity,
			weight_grams, cost_cents, notes, created_on
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::date
		)
		RETURNING `+itemColumns,
		item.ID, userID, item.Name, item.StockNumber, item.Kind, item.Metal, item.Purity,
		item.WeightGrams, item.CostCents, item.Notes, item.CreatedOn,
	)
	stored, err := scanItem(row)
	if err != nil {
		return domain.Item{}, fmt.Errorf("create item: %w", err)
	}
	return stored, nil
}

func (p *Postgres) GetItem(ctx context.Context, userID, id string) (domain.Item, error) {
	item, err := scanItem(p.pool.QueryRow(ctx, `
		SELECT `+itemColumns+`
		FROM items
		WHERE id = $1 AND user_id = $2`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, domain.ErrMissing
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("get item: %w", err)
	}
	return item, nil
}

func (p *Postgres) UpdateItem(ctx context.Context, userID, id string, in domain.Input) (domain.Item, error) {
	normalized, err := domain.Normalize(in)
	if err != nil {
		return domain.Item{}, err
	}
	return p.change(ctx, userID, id, func(current domain.Item) (domain.Item, error) {
		return domain.Edit(current, normalized)
	})
}

func (p *Postgres) MarkSold(ctx context.Context, userID, id, soldOn string, soldCents int, today string) (domain.Item, error) {
	return p.change(ctx, userID, id, func(current domain.Item) (domain.Item, error) {
		return domain.Sell(current, soldOn, soldCents, today)
	})
}

func (p *Postgres) ReturnToStock(ctx context.Context, userID, id string) (domain.Item, error) {
	return p.change(ctx, userID, id, func(current domain.Item) (domain.Item, error) {
		return domain.Unsell(current)
	})
}

func (p *Postgres) DeleteItem(ctx context.Context, userID, id string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM items WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrMissing
	}
	return nil
}

func (p *Postgres) change(ctx context.Context, userID, id string, apply func(domain.Item) (domain.Item, error)) (domain.Item, error) {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return domain.Item{}, err
	}
	defer tx.Rollback(ctx)

	current, err := scanItem(tx.QueryRow(ctx, `
		SELECT `+itemColumns+`
		FROM items
		WHERE id = $1 AND user_id = $2
		FOR UPDATE`, id, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Item{}, domain.ErrMissing
	}
	if err != nil {
		return domain.Item{}, fmt.Errorf("lock item: %w", err)
	}
	next, err := apply(current)
	if err != nil {
		return domain.Item{}, err
	}
	stored, err := scanItem(tx.QueryRow(ctx, `
		UPDATE items SET
			name = $3,
			stock_number = $4,
			kind = $5,
			metal = $6,
			purity = $7,
			weight_grams = $8,
			cost_cents = $9,
			notes = $10,
			sold_on = $11::date,
			sold_cents = $12,
			updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING `+itemColumns,
		id, userID, next.Name, next.StockNumber, next.Kind, next.Metal, next.Purity,
		next.WeightGrams, next.CostCents, next.Notes, dateArg(next.SoldOn), next.SoldCents,
	))
	if err != nil {
		return domain.Item{}, fmt.Errorf("update item: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Item{}, err
	}
	return stored, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanItem(row scanner) (domain.Item, error) {
	var (
		item      domain.Item
		createdOn time.Time
		soldOn    *time.Time
		soldCents *int64
	)
	err := row.Scan(
		&item.ID, &item.Name, &item.StockNumber, &item.Kind, &item.Metal, &item.Purity,
		&item.WeightGrams, &item.CostCents, &item.Notes, &createdOn, &soldOn, &soldCents,
	)
	if err != nil {
		return domain.Item{}, err
	}
	item.CreatedOn = dayString(createdOn)
	if soldOn != nil {
		value := dayString(*soldOn)
		item.SoldOn = &value
	}
	if soldCents != nil {
		value := int(*soldCents)
		item.SoldCents = &value
	}
	return item, nil
}

func dayString(value time.Time) string {
	year, month, day := value.Date()
	return fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
}

func dateArg(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
