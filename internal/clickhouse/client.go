package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// Client provides methods to write analytics batches to ClickHouse tables.
type Client interface {
	InsertUsageBatch(ctx context.Context, records []UsageEvent) error
	InsertBillingEventBatch(ctx context.Context, events []BillingEventLog) error
}

type chClient struct {
	cfg    Config
	logger *zap.Logger
	mu     sync.Mutex
	conn   clickhouse.Conn
}

// NewClient constructs a ClickHouse client using the provided configuration.
func NewClient(cfg Config, logger *zap.Logger) (Client, error) {
	if cfg.DSN == "" {
		return nil, errors.New("CLICKHOUSE_DSN is required")
	}
	c := &chClient{cfg: cfg, logger: logger.Named("clickhouse.client")}
	if err := c.connect(context.Background()); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *chClient) connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	opts, err := clickhouse.ParseDSN(c.cfg.DSN)
	if err != nil {
		return fmt.Errorf("parse dsn: %w", err)
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return fmt.Errorf("open clickhouse: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return fmt.Errorf("ping clickhouse: %w", err)
	}
	c.conn = conn
	return nil
}

func (c *chClient) ensureConn(ctx context.Context) error {
	c.mu.Lock()
	conn := c.conn
	c.mu.Unlock()

	if conn == nil {
		return c.connect(ctx)
	}
	if err := conn.Ping(ctx); err != nil {
		c.logger.Warn("clickhouse ping failed, reconnecting", zap.Error(err))
		return c.connect(ctx)
	}
	return nil
}

func (c *chClient) InsertUsageBatch(ctx context.Context, records []UsageEvent) error {
	if len(records) == 0 {
		return nil
	}
	if err := c.ensureConn(ctx); err != nil {
		return err
	}

	for start := 0; start < len(records); start += c.cfg.BatchSize {
		end := start + c.cfg.BatchSize
		if end > len(records) {
			end = len(records)
		}
		if err := c.insertUsageChunk(ctx, records[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func (c *chClient) insertUsageChunk(ctx context.Context, records []UsageEvent) error {
	c.mu.Lock()
	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO usage_events (tenant_id, customer_id, subscription_id, meter_code, value, recorded_at, idempotency_key, metadata) VALUES (?,?,?,?,?,?,?,?)")
	c.mu.Unlock()
	if err != nil {
		return err
	}
	for _, r := range records {
		if err := batch.Append(
			r.TenantID,
			r.CustomerID,
			r.SubscriptionID,
			r.MeterCode,
			r.Value,
			r.RecordedAt,
			r.IdempotencyKey,
			r.Metadata,
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

func (c *chClient) InsertBillingEventBatch(ctx context.Context, events []BillingEventLog) error {
	if len(events) == 0 {
		return nil
	}
	if err := c.ensureConn(ctx); err != nil {
		return err
	}

	for start := 0; start < len(events); start += c.cfg.BatchSize {
		end := start + c.cfg.BatchSize
		if end > len(events) {
			end = len(events)
		}
		if err := c.insertBillingChunk(ctx, events[start:end]); err != nil {
			return err
		}
	}
	return nil
}

func (c *chClient) insertBillingChunk(ctx context.Context, events []BillingEventLog) error {
	c.mu.Lock()
	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO billing_event_log (event_id, tenant_id, event_type, payload, created_at) VALUES (?,?,?,?,?)")
	c.mu.Unlock()
	if err != nil {
		return err
	}
	for _, evt := range events {
		if err := batch.Append(
			evt.EventID,
			evt.TenantID,
			evt.EventType,
			evt.Payload,
			evt.CreatedAt,
		); err != nil {
			return err
		}
	}
	return batch.Send()
}

// retryWithBackoff retries the provided fn with exponential backoff.
func retryWithBackoff(ctx context.Context, attempts int, base time.Duration, fn func() error) error {
	if attempts < 1 {
		attempts = 1
	}
	delay := base
	for i := 0; i < attempts; i++ {
		if err := fn(); err != nil {
			if i == attempts-1 {
				return err
			}
			select {
			case <-time.After(delay):
				delay *= 2
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	}
	return nil
}
