package idempotency

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smallbiznis/corebilling/internal/headers"
	"github.com/stretchr/testify/assert"
)

type stubRepo struct {
	getFunc    func(ctx context.Context, tenantID, key string) (*Record, error)
	insertFunc func(ctx context.Context, tenantID, key, requestHash string) error
	markFunc   func(ctx context.Context, tenantID, key string, response []byte) error
}

func (s *stubRepo) Get(ctx context.Context, tenantID, key string) (*Record, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, tenantID, key)
	}
	return nil, assert.AnError
}

func (s *stubRepo) InsertProcessing(ctx context.Context, tenantID, key, requestHash string) error {
	if s.insertFunc != nil {
		return s.insertFunc(ctx, tenantID, key, requestHash)
	}
	return nil
}

func (s *stubRepo) MarkCompleted(ctx context.Context, tenantID, key string, response []byte) error {
	if s.markFunc != nil {
		return s.markFunc(ctx, tenantID, key, response)
	}
	return nil
}

type stubCache struct {
	getFunc    func(ctx context.Context, tenantID, key string) (string, error)
	setFunc    func(ctx context.Context, tenantID, key, hash string, ttl time.Duration) error
	deleteFunc func(ctx context.Context, tenantID, key string) error
}

func (s *stubCache) GetHash(ctx context.Context, tenantID, key string) (string, error) {
	if s.getFunc != nil {
		return s.getFunc(ctx, tenantID, key)
	}
	return "", nil
}

func (s *stubCache) SetHash(ctx context.Context, tenantID, key, hash string, ttl time.Duration) error {
	if s.setFunc != nil {
		return s.setFunc(ctx, tenantID, key, hash, ttl)
	}
	return nil
}

func (s *stubCache) Delete(ctx context.Context, tenantID, key string) error {
	if s.deleteFunc != nil {
		return s.deleteFunc(ctx, tenantID, key)
	}
	return nil
}

func TestMiddlewareNoKey(t *testing.T) {
	svc := NewService(&stubRepo{}, &stubCache{})

	handlerCalled := false
	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"data":"x"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, handlerCalled)
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestMiddlewareHeaderOverridesBody(t *testing.T) {
	repo := &stubRepo{}
	cache := &stubCache{}
	svc := NewService(repo, cache)

	body := []byte(`{"idempotency_key":"body-key"}`)
	hash := ComputeHash(body)

	repo.insertFunc = func(ctx context.Context, tenantID, key, requestHash string) error {
		assert.Equal(t, "header-key", key)
		assert.Equal(t, hash, requestHash)
		return nil
	}
	repo.getFunc = func(ctx context.Context, tenantID, key string) (*Record, error) {
		return nil, assert.AnError
	}
	repo.markFunc = func(ctx context.Context, tenantID, key string, response []byte) error {
		assert.Equal(t, "header-key", key)
		assert.JSONEq(t, `{"ok":true}`, string(response))
		return nil
	}
	cache.setFunc = func(ctx context.Context, tenantID, key, h string, ttl time.Duration) error {
		assert.Equal(t, "header-key", key)
		assert.Equal(t, hash, h)
		return nil
	}
	cache.getFunc = func(ctx context.Context, tenantID, key string) (string, error) {
		assert.Equal(t, "header-key", key)
		return "", nil
	}
	cache.deleteFunc = func(ctx context.Context, tenantID, key string) error {
		assert.Equal(t, "header-key", key)
		return nil
	}

	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	req.Header.Set(headers.HeaderIdempotency, "header-key")
	req.Header.Set(headers.HeaderTenantID, "tenant")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
}

func TestMiddlewareReplayedRequest(t *testing.T) {
	hash := ComputeHash([]byte(`{"idempotency_key":"abc"}`))
	repo := &stubRepo{getFunc: func(ctx context.Context, tenantID, key string) (*Record, error) {
		return &Record{Status: StatusProcessing, RequestHash: hash}, nil
	}}
	cache := &stubCache{getFunc: func(ctx context.Context, tenantID, key string) (string, error) {
		return hash, nil
	}}
	svc := NewService(repo, cache)

	handlerCalled := false
	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"idempotency_key":"abc"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusAccepted, rr.Code)
}

func TestMiddlewareNewRequestCompletes(t *testing.T) {
	var completeCalled bool
	hash := ComputeHash([]byte(`{"idempotency_key":"abc"}`))
	repo := &stubRepo{
		getFunc:    func(ctx context.Context, tenantID, key string) (*Record, error) { return nil, assert.AnError },
		insertFunc: func(ctx context.Context, tenantID, key, requestHash string) error { return nil },
		markFunc: func(ctx context.Context, tenantID, key string, response []byte) error {
			completeCalled = true
			assert.JSONEq(t, `{"result":"ok"}`, string(response))
			return nil
		},
	}
	cache := &stubCache{
		getFunc: func(ctx context.Context, tenantID, key string) (string, error) { return "", nil },
		setFunc: func(ctx context.Context, tenantID, key, h string, ttl time.Duration) error {
			assert.Equal(t, hash, h)
			return nil
		},
		deleteFunc: func(ctx context.Context, tenantID, key string) error { return nil },
	}
	svc := NewService(repo, cache)

	handlerCalled := false
	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"ok"}`))
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"idempotency_key":"abc"}`))
	req.Header.Set(headers.HeaderTenantID, "tenant")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, handlerCalled)
	assert.True(t, completeCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddlewarePropagatesPanic(t *testing.T) {
	hash := ComputeHash([]byte(`{"idempotency_key":"abc"}`))
	repo := &stubRepo{getFunc: func(ctx context.Context, tenantID, key string) (*Record, error) { return nil, assert.AnError }, insertFunc: func(ctx context.Context, tenantID, key, requestHash string) error { return nil }}
	cache := &stubCache{getFunc: func(ctx context.Context, tenantID, key string) (string, error) { return "", nil }, setFunc: func(ctx context.Context, tenantID, key, h string, ttl time.Duration) error {
		assert.Equal(t, hash, h)
		return nil
	}}
	svc := NewService(repo, cache)

	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"idempotency_key":"abc"}`))
	rr := httptest.NewRecorder()

	assert.Panics(t, func() { handler.ServeHTTP(rr, req) })
}

func TestMiddlewareReturnsCompletedResponse(t *testing.T) {
	record := &Record{Status: StatusCompleted, Response: []byte(`{"cached":true}`), RequestHash: ComputeHash([]byte(`{"idempotency_key":"abc"}`))}
	repo := &stubRepo{getFunc: func(ctx context.Context, tenantID, key string) (*Record, error) { return record, nil }}
	cache := &stubCache{getFunc: func(ctx context.Context, tenantID, key string) (string, error) { return "", nil }}
	svc := NewService(repo, cache)

	handlerCalled := false
	handler := Middleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"idempotency_key":"abc"}`))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, string(record.Response), rr.Body.String())
}

func TestResponseRecorderCapturesBody(t *testing.T) {
	var recorder responseRecorder
	rr := httptest.NewRecorder()
	recorder.ResponseWriter = rr

	recorder.Write([]byte("hello"))
	recorder.Write([]byte(" world"))

	assert.Equal(t, "hello world", recorder.body.String())
}
