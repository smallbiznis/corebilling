package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestServiceCreateErrors(t *testing.T) {
	repo := NewTestRepository()
	repo.FailCreate = true
	svc := NewService(repo, zap.NewNop())

	err := svc.Create(context.Background(), Subscription{ID: "sub-1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceListPagination(t *testing.T) {
	repo := NewTestRepository()
	for i := 0; i < 3; i++ {
		repo.Subs[fmt.Sprintf("sub-%d", i)] = Subscription{
			ID:        fmt.Sprintf("sub-%d", i),
			TenantID:  "tenant",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Minute),
		}
	}
	svc := NewService(repo, zap.NewNop())

	items, hasMore, err := svc.List(context.Background(), ListSubscriptionsFilter{
		TenantID: "tenant",
		Limit:    2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items got %d", len(items))
	}
	if !hasMore {
		t.Fatal("expected next page")
	}
}
