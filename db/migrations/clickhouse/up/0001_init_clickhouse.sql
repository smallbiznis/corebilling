-- ===============================
-- 1. USAGE EVENTS (raw usage)
-- ===============================
CREATE TABLE IF NOT EXISTS usage_events
(
    tenant_id String,
    customer_id String,
    subscription_id String,
    meter_code String,
    value Float64,
    recorded_at DateTime,
    day Date DEFAULT toDate(recorded_at),
    idempotency_key String,
    metadata String,
    ingestion_ts DateTime DEFAULT now()
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(day)
ORDER BY (tenant_id, meter_code, day);

-- ===============================
-- 2. BILLING EVENT LOG (BEaaS event pipeline)
-- ===============================
CREATE TABLE IF NOT EXISTS billing_event_log
(
    event_id String,
    tenant_id String,
    event_type String,
    payload String,
    created_at DateTime,
    day Date DEFAULT toDate(created_at)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(day)
ORDER BY (tenant_id, event_type, day);

-- ===============================
-- 3. INVOICE ROLLUPS (aggregated invoice analytics)
-- ===============================
CREATE TABLE IF NOT EXISTS invoice_rollups
(
    tenant_id String,
    invoice_id String,
    subscription_id String,
    cycle_start DateTime,
    cycle_end DateTime,
    total_amount Float64,
    usage_amount Float64,
    tax_amount Float64,
    created_at DateTime,
    day Date DEFAULT toDate(created_at)
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(day)
ORDER BY (tenant_id, invoice_id);
