package domain

import (
	"context"
	"errors"
)

// TestRepository is an in-memory repo for tests.
type TestRepository struct {
	Subs       map[string]Subscription
	FailCreate bool
	FailGet    bool
	FailList   bool
	FailUpdate bool
}

// NewTestRepository creates a fresh test repo.
func NewTestRepository() *TestRepository {
	return &TestRepository{
		Subs: make(map[string]Subscription),
	}
}

func (r *TestRepository) Create(ctx context.Context, sub Subscription) error {
	if r.FailCreate {
		return errors.New("create error")
	}
	r.Subs[sub.ID] = sub
	return nil
}

func (r *TestRepository) GetByID(ctx context.Context, id string) (Subscription, error) {
	if r.FailGet {
		return Subscription{}, errors.New("get error")
	}
	sub, ok := r.Subs[id]
	if !ok {
		return Subscription{}, errors.New("not found")
	}
	return sub, nil
}

func (r *TestRepository) List(ctx context.Context, filter ListSubscriptionsFilter) ([]Subscription, bool, error) {
	if r.FailList {
		return nil, false, errors.New("list error")
	}
	var matches []Subscription
	for _, sub := range r.Subs {
		if filter.TenantID != "" && sub.TenantID != filter.TenantID {
			continue
		}
		if filter.CustomerID != "" && sub.CustomerID != filter.CustomerID {
			continue
		}
		matches = append(matches, sub)
	}
	if filter.Limit > 0 && len(matches) > filter.Limit {
		return matches[:filter.Limit], true, nil
	}
	return matches, false, nil
}

func (r *TestRepository) Update(ctx context.Context, sub Subscription) error {
	if r.FailUpdate {
		return errors.New("update error")
	}
	r.Subs[sub.ID] = sub
	return nil
}
