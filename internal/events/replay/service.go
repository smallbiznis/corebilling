package replay

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	"github.com/smallbiznis/corebilling/internal/events/replay/repository"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"google.golang.org/protobuf/types/known/structpb"
)

// Service exposes replay capabilities for stored events.
type Service interface {
	ReplayEvent(ctx context.Context, eventID string) error
	ReplayAllForTenant(ctx context.Context, tenantID string) error
	ReplayByType(ctx context.Context, eventType string) error
	ReplayAdvanced(ctx context.Context, opts ReplayOptions) error
}

type service struct {
	repo       repository.Repository
	outboxRepo outbox.OutboxRepository
	logger     *zap.Logger
}

// NewService constructs a replay service with sane defaults.
func NewService(repo repository.Repository, outboxRepo outbox.OutboxRepository, logger *zap.Logger) Service {
	return &service{
		repo:       repo,
		outboxRepo: outboxRepo,
		logger:     logger.Named("event_replay.service"),
	}
}

// ReplayEvent replays a single event by ID.
func (s *service) ReplayEvent(ctx context.Context, eventID string) error {
	env, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return err
	}

	s.logger.Info("replaying single event", zap.String("event_id", env.ID), zap.String("tenant_id", env.TenantID), zap.String("event_type", env.EventType), zap.String("source", "ReplayEvent"))
	return s.publish(ctx, env, false, "ReplayEvent")
}

// ReplayAllForTenant replays all events for a tenant.
func (s *service) ReplayAllForTenant(ctx context.Context, tenantID string) error {
	events, err := s.repo.ListEventsForTenant(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, env := range events {
		if err := s.publish(ctx, env, true, "ReplayAllForTenant"); err != nil {
			return err
		}
	}
	return nil
}

// ReplayByType replays all events of the provided type.
func (s *service) ReplayByType(ctx context.Context, eventType string) error {
	events, err := s.repo.ListEventsByType(ctx, eventType)
	if err != nil {
		return err
	}
	for _, env := range events {
		if err := s.publish(ctx, env, true, "ReplayByType"); err != nil {
			return err
		}
	}
	return nil
}

// ReplayAdvanced replays using flexible filters and supports dry-run mode.
func (s *service) ReplayAdvanced(ctx context.Context, opts ReplayOptions) error {
	events, err := s.repo.ListEventsByFilters(ctx, opts.TenantID, opts.EventType, opts.Since, opts.Until)
	if err != nil {
		return err
	}

	for _, env := range events {
		if opts.DryRun {
			s.logger.Info("dry-run replay", zap.String("event_id", env.ID), zap.String("tenant_id", env.TenantID), zap.String("event_type", env.EventType))
			continue
		}
		if err := s.publish(ctx, env, true, "ReplayAdvanced"); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) publish(ctx context.Context, env repository.EventEnvelope, batch bool, source string) error {
	if env.Event == nil {
		return errors.New("event payload missing")
	}

	ensureMetadata(env.Event)
	env.Event.Metadata.Fields[ReplayFlagKey] = structpb.NewBoolValue(true)
	env.Event.Metadata.Fields[ReplayTimestampKey] = structpb.NewStringValue(time.Now().UTC().Format(time.RFC3339))
	env.Event.Metadata.Fields[ReplaySourceKey] = structpb.NewStringValue(source)
	if batch {
		env.Event.Metadata.Fields[ReplayBatchKey] = structpb.NewBoolValue(true)
	}

	log := ctxlogger.FromContext(ctx).With(
		zap.String("event_id", env.ID),
		zap.String("tenant_id", env.TenantID),
		zap.String("event_type", env.EventType),
		zap.String("replay_source", source),
	)

	log.Info("publishing replayed event")

	return s.outboxRepo.InsertOutboxEvent(ctx, &outbox.OutboxEvent{
		ID:         env.ID,
		Subject:    env.Subject,
		TenantID:   env.TenantID,
		ResourceID: env.ResourceID,
		Event:      env.Event,
	})
}

func ensureMetadata(evt *events.Event) {
	if evt.Metadata == nil {
		evt.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
	}
	if evt.Metadata.Fields == nil {
		evt.Metadata.Fields = map[string]*structpb.Value{}
	}
}
