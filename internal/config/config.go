package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds application configuration.
type Config struct {
	DatabaseURL              string
	ServiceName              string
	ServiceVersion           string
	Environment              string
	MigrationsRoot           string
	EnabledMigrationServices []string
	OTLPEndpoint             string
}

// Load loads configuration from environment variables and .env file.
func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		DatabaseURL:              getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/corebilling?sslmode=disable"),
		ServiceName:              getenv("SERVICE_NAME", "corebilling"),
		ServiceVersion:           getenv("SERVICE_VERSION", "0.1.0"),
		Environment:              getenv("ENVIRONMENT", "development"),
		MigrationsRoot:           getenv("MIGRATIONS_ROOT", "db/migrations"),
		EnabledMigrationServices: parseServices(getenv("ENABLED_MIGRATION_SERVICES", "billing,pricing,subscription,usage,rating,invoice")),
		OTLPEndpoint:             getenv("OTLP_ENDPOINT", "localhost:4317"),
	}
	return cfg
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseServices(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		log.Println("no services enabled for migration")
	}
	return out
}
