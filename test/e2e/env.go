package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/smallbiznis/corebilling/internal/app"
	"github.com/testcontainers/testcontainers-go/modules/nats"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/fx"
)

// testEnv encapsulates the shared infrastructure for E2E tests.
type testEnv struct {
	PGContainer   *postgres.PostgresContainer
	NATSContainer *nats.NATSContainer
	App           *fx.App
	GRPCAddr      string
	HTTPAddr      string
}

// setupEnv boots postgres, nats and the corebilling app using Fx.
func setupEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	pg, err := postgres.RunContainer(ctx,
		postgres.WithDatabase("postgres"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	natsC, err := nats.RunContainer(ctx)
	if err != nil {
		pg.Terminate(ctx)
		t.Fatalf("start nats: %v", err)
	}

	dsn, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pg.Terminate(ctx)
		natsC.Terminate(ctx)
		t.Fatalf("dsn: %v", err)
	}

	// Ensure ports are known for clients inside tests.
	grpcAddr := ":50052"
	httpAddr := ":8080"

	// Configure environment for the application.
	os.Setenv("DATABASE_URL", dsn)
	os.Setenv("EVENT_BUS_PROVIDER", "nats")
	natsURL, _ := natsC.ConnectionString(ctx)
	os.Setenv("NATS_URL", natsURL)
	os.Setenv("NATS_USERNAME", "")
	os.Setenv("NATS_PASSWORD", "")

	app := appInstance()
	startCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := app.Start(startCtx); err != nil {
		pg.Terminate(ctx)
		natsC.Terminate(ctx)
		t.Fatalf("start app: %v", err)
	}

	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		_ = app.Stop(stopCtx)
		pg.Terminate(context.Background())
		natsC.Terminate(context.Background())
	})

	return &testEnv{
		PGContainer:   pg,
		NATSContainer: natsC,
		App:           app,
		GRPCAddr:      grpcAddr,
		HTTPAddr:      httpAddr,
	}
}

func appInstance() *fx.App {
	return app.Billing()
}
