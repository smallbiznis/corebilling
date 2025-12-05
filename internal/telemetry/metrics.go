package telemetry

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics exposes Prometheus observability primitives for BEaaS.
type Metrics struct {
	apiRequests         *prometheus.CounterVec
	apiDuration         *prometheus.HistogramVec
	outboxDispatch      *prometheus.CounterVec
	outboxDispatchTime  *prometheus.HistogramVec
	outboxBacklog       prometheus.Gauge
	handlerDuration     *prometheus.HistogramVec
	handlerErrors       *prometheus.CounterVec
	webhookDeliveries   *prometheus.CounterVec
	webhookDuration     *prometheus.HistogramVec
	webhookBacklogGauge prometheus.Gauge
}

// NewMetrics registers and returns Prometheus metrics for telemetry.
func NewMetrics() *Metrics {
	apiRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "corebilling_api_requests_total",
		Help: "Counts API requests by method, status, and tenant.",
	}, []string{"method", "status", "tenant"})

	apiDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "corebilling_api_duration_seconds",
		Help:    "API request latency per method/tenant.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "tenant"})

	outboxDispatch := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "corebilling_outbox_dispatch_total",
		Help: "Counts dispatcher batches by status.",
	}, []string{"status"})

	outboxDispatchTime := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "corebilling_outbox_dispatch_duration_seconds",
		Help:    "Dispatcher batch durations.",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	outboxBacklog := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "corebilling_outbox_backlog",
		Help: "Number of pending events in the outbox.",
	})

	handlerDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "corebilling_event_handler_duration_seconds",
		Help:    "Event handler durations by subject.",
		Buckets: prometheus.DefBuckets,
	}, []string{"subject", "tenant", "status"})

	handlerErrors := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "corebilling_event_handler_errors_total",
		Help: "Counts handler errors by subject and tenant.",
	}, []string{"subject", "tenant"})

	webhookDeliveries := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "corebilling_webhook_delivery_total",
		Help: "Webhook delivery outcomes.",
	}, []string{"status", "tenant"})

	webhookDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "corebilling_webhook_delivery_duration_seconds",
		Help:    "Webhook delivery roundtrip latency.",
		Buckets: prometheus.DefBuckets,
	}, []string{"tenant"})

	webhookBacklog := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "corebilling_webhook_backlog",
		Help: "Number of pending webhook attempts.",
	})

	prometheus.MustRegister(
		apiRequests,
		apiDuration,
		outboxDispatch,
		outboxDispatchTime,
		outboxBacklog,
		handlerDuration,
		handlerErrors,
		webhookDeliveries,
		webhookDuration,
		webhookBacklog,
	)

	return &Metrics{
		apiRequests:         apiRequests,
		apiDuration:         apiDuration,
		outboxDispatch:      outboxDispatch,
		outboxDispatchTime:  outboxDispatchTime,
		outboxBacklog:       outboxBacklog,
		handlerDuration:     handlerDuration,
		handlerErrors:       handlerErrors,
		webhookDeliveries:   webhookDeliveries,
		webhookDuration:     webhookDuration,
		webhookBacklogGauge: webhookBacklog,
	}
}

// ObserveAPIRequest records an API request and latency.
func (m *Metrics) ObserveAPIRequest(method, status, tenant string, duration time.Duration) {
	if m == nil {
		return
	}
	tenantLabel := sanitizeTenant(tenant)
	methodLabel := sanitizeLabel(method)
	m.apiRequests.WithLabelValues(methodLabel, status, tenantLabel).Inc()
	m.apiDuration.WithLabelValues(methodLabel, tenantLabel).Observe(duration.Seconds())
}

// RecordOutboxBatch registers dispatch batch metrics.
func (m *Metrics) RecordOutboxBatch(status string, count int, duration time.Duration) {
	if m == nil {
		return
	}
	m.outboxDispatch.WithLabelValues(status).Inc()
	m.outboxDispatchTime.WithLabelValues(status).Observe(duration.Seconds())
}

// SetOutboxBacklog updates the backlog gauge.
func (m *Metrics) SetOutboxBacklog(value float64) {
	if m == nil {
		return
	}
	m.outboxBacklog.Set(value)
}

// RecordHandler observes handler invocations.
func (m *Metrics) RecordHandler(subject, tenant, status string, duration time.Duration) {
	if m == nil {
		return
	}
	tenantLabel := sanitizeTenant(tenant)
	m.handlerDuration.WithLabelValues(subject, tenantLabel, status).Observe(duration.Seconds())
	if status != "success" {
		m.handlerErrors.WithLabelValues(subject, tenantLabel).Inc()
	}
}

// RecordWebhookDelivery records webhook delivery metrics.
func (m *Metrics) RecordWebhookDelivery(status, tenant string, duration time.Duration) {
	if m == nil {
		return
	}
	tenantLabel := sanitizeTenant(tenant)
	m.webhookDeliveries.WithLabelValues(status, tenantLabel).Inc()
	m.webhookDuration.WithLabelValues(tenantLabel).Observe(duration.Seconds())
}

// ObserveWebhookBacklog updates the webhook backlog gauge.
func (m *Metrics) ObserveWebhookBacklog(value float64) {
	if m == nil {
		return
	}
	m.webhookBacklogGauge.Set(value)
}

func sanitizeTenant(tenant string) string {
	if tenant == "" {
		return "unknown"
	}
	return tenant
}

func sanitizeLabel(val string) string {
	if val == "" {
		return "unknown"
	}
	return val
}
