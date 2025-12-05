package outbox

import "context"

// DLQService exposes read-only DLQ operations.
type DLQService struct {
	repo DeadLetterRepository
}

// NewDLQService constructs a DLQService.
func NewDLQService(repo DeadLetterRepository) *DLQService {
	return &DLQService{repo: repo}
}

// List returns dead-lettered events honoring bounds.
func (s *DLQService) List(ctx context.Context, limit, offset int32) ([]OutboxEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListDeadLetters(ctx, limit, offset)
}
