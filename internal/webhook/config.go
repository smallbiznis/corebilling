package webhook

import (
	"os"
	"strconv"
	"time"
)

// Config controls webhook worker timing and retries.
type Config struct {
	Interval    time.Duration
	Limit       int32
	MaxRetries  int32
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	HTTPTimeout time.Duration
}

// LoadConfig builds webhook-specific configuration from environment variables.
func LoadConfig() Config {
	return Config{
		Interval:    durationFromEnv("WEBHOOK_WORKER_INTERVAL_SECONDS", 30),
		Limit:       int32(intFromEnv("WEBHOOK_WORKER_LIMIT", 50)),
		MaxRetries:  int32(intFromEnv("WEBHOOK_MAX_RETRIES", 5)),
		BaseDelay:   durationFromEnv("WEBHOOK_BASE_DELAY_SECONDS", 10),
		MaxDelay:    durationFromEnv("WEBHOOK_MAX_DELAY_SECONDS", 300),
		HTTPTimeout: durationFromEnv("WEBHOOK_HTTP_TIMEOUT_SECONDS", 15),
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

func durationFromEnv(key string, def int) time.Duration {
	return time.Duration(intFromEnv(key, def)) * time.Second
}
