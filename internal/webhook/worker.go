package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/smallbiznis/corebilling/internal/headers"
	"github.com/smallbiznis/corebilling/internal/telemetry"
	"github.com/smallbiznis/corebilling/internal/webhook/repository"
	"go.uber.org/zap"
)

const (
	statusSuccess = "SUCCESS"
	statusFailed  = "FAILED"
	statusDLQ     = "DLQ"
)

type Worker struct {
	repo    repository.Repository
	client  *http.Client
	logger  *zap.Logger
	cfg     Config
	rnd     *rand.Rand
	metrics *telemetry.Metrics
}

func NewWorker(repo repository.Repository, cfg Config, client *http.Client, metrics *telemetry.Metrics, logger *zap.Logger) *Worker {
	if client == nil {
		client = &http.Client{Timeout: cfg.HTTPTimeout}
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	return &Worker{
		repo:    repo,
		client:  client,
		logger:  logger.Named("webhook.worker"),
		cfg:     cfg,
		rnd:     rand.New(rand.NewSource(time.Now().UnixNano())),
		metrics: metrics,
	}
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *Worker) process(ctx context.Context) {
	attempts, err := w.repo.ListDueDeliveryAttempts(ctx, w.cfg.Limit)
	if err != nil {
		w.logger.Error("failed to fetch due webhook attempts", zap.Error(err))
		return
	}
	if w.metrics != nil {
		w.metrics.ObserveWebhookBacklog(float64(len(attempts)))
	}
	for _, attempt := range attempts {
		w.deliver(ctx, attempt)
	}
}

func (w *Worker) deliver(ctx context.Context, attempt repository.WebhookDeliveryAttempt) {
	start := time.Now()
	log := w.logger.With(
		zap.String("webhook_id", attempt.WebhookID),
		zap.String("event_id", attempt.EventID),
		zap.String("tenant_id", attempt.TenantID),
		zap.Int32("attempt_no", attempt.AttemptNo),
	)

	webhook, err := w.repo.GetWebhook(ctx, attempt.WebhookID)
	if err != nil {
		log.Error("failed to load webhook", zap.Error(err))
		return
	}

	subject, eventID := decodeEventMeta(attempt.Payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.TargetUrl, bytes.NewReader(attempt.Payload))
	if err != nil {
		log.Error("failed to build http request", zap.Error(err))
		w.handleFailure(ctx, attempt, fmt.Errorf("request build: %w", err), time.Since(start))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headers.HeaderTenantID, attempt.TenantID)
	req.Header.Set(headers.HeaderEventID, eventID)
	if subject != "" {
		req.Header.Set(headers.HeaderEventType, subject)
	}
	if signature := SignPayload(webhook.Secret, attempt.Payload); signature != "" {
		req.Header.Set(headers.HeaderSignature, signature)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		log.Error("http delivery failed", zap.Error(err))
		w.handleFailure(ctx, attempt, err, time.Since(start))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		w.recordSuccess(ctx, attempt, time.Since(start))
		// TODO: increment webhookDeliverySuccess counter with tenant_id, webhook_id.
		return
	}

	msg := fmt.Sprintf("unexpected status %d", resp.StatusCode)
	log.Warn("webhook delivery failed", zap.String("status", msg))
	w.handleFailure(ctx, attempt, errors.New(msg), time.Since(start))
}

func (w *Worker) recordSuccess(ctx context.Context, attempt repository.WebhookDeliveryAttempt, duration time.Duration) {
	if w.metrics != nil {
		w.metrics.RecordWebhookDelivery("success", attempt.TenantID, duration)
	}
	_ = w.repo.UpdateDeliveryAttemptStatus(ctx, repository.UpdateDeliveryAttemptStatusParams{
		ID:        attempt.ID,
		Status:    statusSuccess,
		AttemptNo: attempt.AttemptNo + 1,
		NextRunAt: toTimestamptz(time.Now().UTC()),
		LastError: textFromError(nil),
	})
}

func (w *Worker) handleFailure(ctx context.Context, attempt repository.WebhookDeliveryAttempt, err error, duration time.Duration) {
	log := w.logger.With(
		zap.String("webhook_id", attempt.WebhookID),
		zap.String("event_id", attempt.EventID),
		zap.String("tenant_id", attempt.TenantID),
		zap.Int32("attempt_no", attempt.AttemptNo),
		zap.String("status", statusFailed),
	)
	log.Error("delivery attempt failed", zap.Error(err))

	if w.metrics != nil {
		w.metrics.RecordWebhookDelivery("failed", attempt.TenantID, duration)
	}

	nextAttemptNo := attempt.AttemptNo + 1
	if nextAttemptNo >= w.cfg.MaxRetries {
		if dlqErr := w.repo.MoveToDLQ(ctx, repository.MoveToDLQParams{
			WebhookID: attempt.WebhookID,
			EventID:   attempt.EventID,
			TenantID:  attempt.TenantID,
			Payload:   attempt.Payload,
			Reason:    textFromError(err),
			ID:        attempt.ID,
		}); dlqErr != nil {
			log.Error("failed to move to DLQ", zap.Error(dlqErr))
		}
		if w.metrics != nil {
			w.metrics.RecordWebhookDelivery("dlq", attempt.TenantID, duration)
		}
		// TODO: increment webhookDeliveryDLQ counter.
		return
	}

	nextRun := w.nextRunAt(nextAttemptNo)
	if updateErr := w.repo.UpdateDeliveryAttemptStatus(ctx, repository.UpdateDeliveryAttemptStatusParams{
		ID:        attempt.ID,
		Status:    statusFailed,
		AttemptNo: nextAttemptNo,
		NextRunAt: toTimestamptz(nextRun),
		LastError: textFromError(err),
	}); updateErr != nil {
		log.Error("failed to update failed attempt", zap.Error(updateErr))
	}
}

func (w *Worker) nextRunAt(nextAttemptNo int32) time.Time {
	delay := float64(w.cfg.BaseDelay) * math.Pow(2, float64(nextAttemptNo))
	if w.cfg.MaxDelay > 0 {
		delay = math.Min(delay, float64(w.cfg.MaxDelay))
	}
	jitter := time.Duration(w.rnd.Int63n(int64(delay/2) + 1))
	return time.Now().UTC().Add(time.Duration(delay) + jitter)
}

func decodeEventMeta(payload []byte) (string, string) {
	var meta struct {
		Subject string `json:"subject"`
		ID      string `json:"id"`
	}
	if err := json.Unmarshal(payload, &meta); err != nil {
		return "", ""
	}
	return meta.Subject, meta.ID
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}

func textFromError(err error) pgtype.Text {
	if err == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{
		String: err.Error(),
		Valid:  true,
	}
}
