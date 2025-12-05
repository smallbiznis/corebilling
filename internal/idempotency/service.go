package idempotency

import (
	"context"
	"encoding/json"
	"errors"
)

// Service coordinates Redis and PostgreSQL idempotency layers.
type Service struct {
	repo  Repository
	cache Cache
}

// NewService constructs a Service.
func NewService(repo Repository, cache Cache) *Service {
	return &Service{repo: repo, cache: cache}
}

// Begin evaluates idempotency state for an incoming request.
func (s *Service) Begin(ctx context.Context, tenantID, key string, body []byte) (*Record, bool, error) {
	if key == "" {
		return nil, false, ErrMissingKey
	}
	if s.repo == nil {
		return nil, false, errors.New("idempotency repository not configured")
	}
	requestHash := ComputeHash(body)

	if hash, err := s.cache.GetHash(ctx, tenantID, key); err == nil && hash != "" {
		if hash == requestHash {
			if s.repo != nil {
				if record, err := s.repo.Get(ctx, tenantID, key); err == nil {
					return record, true, nil
				}
			}
			return &Record{TenantID: tenantID, Key: key, RequestHash: requestHash, Status: StatusProcessing}, true, nil
		}
	}

	if s.repo != nil {
		if record, err := s.repo.Get(ctx, tenantID, key); err == nil {
			if record.RequestHash != requestHash {
				return record, true, ErrAlreadyCompleted
			}
			return record, true, nil
		}
	}

	if err := s.repo.InsertProcessing(ctx, tenantID, key, requestHash); err != nil {
		if s.repo != nil {
			if record, getErr := s.repo.Get(ctx, tenantID, key); getErr == nil {
				if record.RequestHash != requestHash {
					return record, true, ErrAlreadyCompleted
				}
				return record, true, nil
			}
		}
		return nil, false, err
	}

	_ = s.cache.SetHash(ctx, tenantID, key, requestHash, defaultCacheTTL)

	return &Record{TenantID: tenantID, Key: key, RequestHash: requestHash, Status: StatusProcessing}, false, nil
}

// Complete finalizes an idempotent request with its response payload.
func (s *Service) Complete(ctx context.Context, tenantID, key string, response interface{}) error {
	if key == "" {
		return ErrMissingKey
	}
	if s.repo == nil {
		return errors.New("idempotency repository not configured")
	}
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	if err := s.repo.MarkCompleted(ctx, tenantID, key, payload); err != nil {
		return err
	}
	_ = s.cache.Delete(ctx, tenantID, key)
	return nil
}

// IsAlreadyCompleted returns true when the error represents a completed request conflict.
func IsAlreadyCompleted(err error) bool {
	return errors.Is(err, ErrAlreadyCompleted)
}
