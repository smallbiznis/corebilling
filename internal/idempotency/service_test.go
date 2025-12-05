package idempotency_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	idempotency "github.com/smallbiznis/corebilling/internal/idempotency"
	"github.com/smallbiznis/corebilling/internal/idempotency/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceBeginCacheHitSameHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)
	repo := mocks.NewMockRepository(ctrl)
	svc := idempotency.NewService(repo, cache)

	body := []byte(`{"foo":"bar"}`)
	hash := idempotency.ComputeHash(body)
	record := &idempotency.Record{TenantID: "t1", Key: "k1", RequestHash: hash, Status: idempotency.StatusCompleted}

	cache.EXPECT().GetHash(gomock.Any(), "t1", "k1").Return(hash, nil)
	repo.EXPECT().Get(gomock.Any(), "t1", "k1").Return(record, nil)

	got, existing, err := svc.Begin(context.Background(), "t1", "k1", body)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, record, got)
}

func TestServiceBeginCacheHitDifferentHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)
	repo := mocks.NewMockRepository(ctrl)
	svc := idempotency.NewService(repo, cache)

	body := []byte(`{"foo":"bar"}`)
	hash := idempotency.ComputeHash(body)

	cache.EXPECT().GetHash(gomock.Any(), "t1", "k1").Return("different", nil)
	repo.EXPECT().Get(gomock.Any(), "t1", "k1").Return(nil, errors.New("not found"))
	repo.EXPECT().InsertProcessing(gomock.Any(), "t1", "k1", hash).Return(nil)
	cache.EXPECT().SetHash(gomock.Any(), "t1", "k1", hash, gomock.Any())

	record, existing, err := svc.Begin(context.Background(), "t1", "k1", body)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, idempotency.StatusProcessing, record.Status)
	require.Equal(t, hash, record.RequestHash)
}

func TestServiceBeginPostgresCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)
	repo := mocks.NewMockRepository(ctrl)
	svc := idempotency.NewService(repo, cache)

	body := []byte(`{"foo":"bar"}`)
	record := &idempotency.Record{TenantID: "t1", Key: "k1", RequestHash: idempotency.ComputeHash(body), Status: idempotency.StatusCompleted}

	cache.EXPECT().GetHash(gomock.Any(), "t1", "k1").Return("", nil)
	repo.EXPECT().Get(gomock.Any(), "t1", "k1").Return(record, nil)

	got, existing, err := svc.Begin(context.Background(), "t1", "k1", body)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, record, got)
}

func TestServiceBeginPostgresProcessing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)
	repo := mocks.NewMockRepository(ctrl)
	svc := idempotency.NewService(repo, cache)

	body := []byte(`{"foo":"bar"}`)
	record := &idempotency.Record{TenantID: "t1", Key: "k1", RequestHash: idempotency.ComputeHash(body), Status: idempotency.StatusProcessing}

	cache.EXPECT().GetHash(gomock.Any(), "t1", "k1").Return("", nil)
	repo.EXPECT().Get(gomock.Any(), "t1", "k1").Return(record, nil)

	got, existing, err := svc.Begin(context.Background(), "t1", "k1", body)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, record, got)
}

func TestServiceBeginNewRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)
	repo := mocks.NewMockRepository(ctrl)
	svc := idempotency.NewService(repo, cache)

	body := []byte(`{"foo":"bar"}`)
	hash := idempotency.ComputeHash(body)

	cache.EXPECT().GetHash(gomock.Any(), "t1", "k1").Return("", nil)
	repo.EXPECT().Get(gomock.Any(), "t1", "k1").Return(nil, errors.New("not found"))
	repo.EXPECT().InsertProcessing(gomock.Any(), "t1", "k1", hash).Return(nil)
	cache.EXPECT().SetHash(gomock.Any(), "t1", "k1", hash, gomock.Any())

	record, existing, err := svc.Begin(context.Background(), "t1", "k1", body)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, idempotency.StatusProcessing, record.Status)
	require.Equal(t, hash, record.RequestHash)
}

func TestServiceComplete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cache := mocks.NewMockCache(ctrl)
	repo := mocks.NewMockRepository(ctrl)
	svc := idempotency.NewService(repo, cache)

	payload := map[string]bool{"ok": true}

	repo.EXPECT().MarkCompleted(gomock.Any(), "tenant", "key", gomock.Any()).DoAndReturn(func(_ context.Context, _ string, _ string, b []byte) error {
		assert.JSONEq(t, `{"ok":true}`, string(b))
		return nil
	})
	cache.EXPECT().Delete(gomock.Any(), "tenant", "key")

	err := svc.Complete(context.Background(), "tenant", "key", payload)
	require.NoError(t, err)
}
