package outbox

import (
	"os"
	"strconv"
)

// Config captures runtime configuration for the outbox repository.
type Config struct {
	ShardID    int
	ShardTotal int
}

// NewConfigFromEnv builds Config using SHARD_ID/SHARD_TOTAL. Defaults to a
// single shard when variables are unset or invalid.
func NewConfigFromEnv() Config {
	id := parseEnvInt("SHARD_ID", 0)
	total := parseEnvInt("SHARD_TOTAL", 1)
	return Config{ShardID: id, ShardTotal: total}
}

func parseEnvInt(key string, def int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return val
}
