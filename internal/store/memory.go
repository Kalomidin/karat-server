package store

import (
	"context"
	"sync"

	"karat-server/internal/domain"
)

// Memory is an in-process stock book used by tests.
type Memory struct {
	mu    sync.Mutex
	users map[string]domain.User
	items map[string][]domain.Item
	owner map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		users: map[string]domain.User{},
		items: map[string][]domain.Item{},
		owner: map[string]string{},
	}
}

func (m *Memory) Ping(context.Context) error { return nil }

func (m *Memory) UpsertUser(_ context.Context, user domain.User) (domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.GoogleID] = user
	return user, nil
}

func (m *Memory) ListItems(_ context.Context, userID string) ([]domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := m.items[userID]
	out := make([]domain.Item, len(items))
	for i, item := range items {
		out[i] = domain.Clone(item)
	}
	return out, nil
}

func (m *Memory) CreateItem(_ context.Context, userID string, in domain.Input, today string) (domain.Item, error) {
	id, err := newID()
	if err != nil {
		return domain.Item{}, err
	}
	item, err := domain.NewItem(id, today, in)
	if err != nil {
		return domain.Item{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[userID] = append([]domain.Item{domain.Clone(item)}, m.items[userID]...)
	m.owner[item.ID] = userID
	return domain.Clone(item), nil
}

func (m *Memory) GetItem(_ context.Context, userID, id string) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	item, _, ok := m.find(userID, id)
	if !ok {
		return domain.Item{}, domain.ErrMissing
	}
	return domain.Clone(item), nil
}

func (m *Memory) UpdateItem(_ context.Context, userID, id string, in domain.Input) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, index, ok := m.find(userID, id)
	if !ok {
		return domain.Item{}, domain.ErrMissing
	}
	next, err := domain.Edit(current, in)
	if err != nil {
		return domain.Item{}, err
	}
	m.items[userID][index] = domain.Clone(next)
	return domain.Clone(next), nil
}

func (m *Memory) MarkSold(_ context.Context, userID, id, soldOn string, soldCents int, today string) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, index, ok := m.find(userID, id)
	if !ok {
		return domain.Item{}, domain.ErrMissing
	}
	next, err := domain.Sell(current, soldOn, soldCents, today)
	if err != nil {
		return domain.Item{}, err
	}
	m.items[userID][index] = domain.Clone(next)
	return domain.Clone(next), nil
}

func (m *Memory) ReturnToStock(_ context.Context, userID, id string) (domain.Item, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	current, index, ok := m.find(userID, id)
	if !ok {
		return domain.Item{}, domain.ErrMissing
	}
	next, err := domain.Unsell(current)
	if err != nil {
		return domain.Item{}, err
	}
	m.items[userID][index] = domain.Clone(next)
	return domain.Clone(next), nil
}

func (m *Memory) DeleteItem(_ context.Context, userID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := m.items[userID]
	for i, item := range items {
		if item.ID != id {
			continue
		}
		m.items[userID] = append(items[:i], items[i+1:]...)
		delete(m.owner, id)
		return nil
	}
	return domain.ErrMissing
}

func (m *Memory) find(userID, id string) (domain.Item, int, bool) {
	if m.owner[id] != userID {
		return domain.Item{}, 0, false
	}
	for i, item := range m.items[userID] {
		if item.ID == id {
			return item, i, true
		}
	}
	return domain.Item{}, 0, false
}
