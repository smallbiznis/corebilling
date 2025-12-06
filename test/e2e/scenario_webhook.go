package e2e

import (
	"testing"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func scenarioWebhook(t *testing.T, env *testEnv, client *apiClient) {
	t.Helper()
	ctx, cancel := newTimeoutCtx()
	defer cancel()

	hc := healthpb.NewHealthClient(client.conn)
	if _, err := hc.Check(ctx, &healthpb.HealthCheckRequest{}); err != nil {
		t.Fatalf("health check failed: %v", err)
	}
}
