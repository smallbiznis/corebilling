package clickhouse

import (
	"os"
	"strconv"
	"time"
)

// Config captures ClickHouse writer configuration.
type Config struct {
	DSN           string
	BatchSize     int
	FlushInterval time.Duration
}

// LoadConfig builds configuration from environment variables with sensible defaults.
func LoadConfig() Config {
	return Config{
		DSN:           os.Getenv("CLICKHOUSE_DSN"),
		BatchSize:     intFromEnv("CLICKHOUSE_BATCH_SIZE", 500),
		FlushInterval: time.Duration(intFromEnv("CLICKHOUSE_FLUSH_INTERVAL_MS", 1000)) * time.Millisecond,
	}
}

func intFromEnv(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return def
}
