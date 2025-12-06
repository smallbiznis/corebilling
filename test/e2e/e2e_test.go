package e2e

import (
	"context"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	env := setupEnv(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := newAPIClient(ctx, env.GRPCAddr)
	if err != nil {
		t.Fatalf("dial grpc: %v", err)
	}
	defer client.close()

	t.Run("subscription lifecycle", func(t *testing.T) {
		scenarioSubscriptionLifecycle(t, env, client)
	})
	t.Run("metering", func(t *testing.T) {
		scenarioMetering(t, env, client)
	})
	t.Run("rating", func(t *testing.T) {
		scenarioRating(t, env, client)
	})
	t.Run("invoice", func(t *testing.T) {
		scenarioInvoice(t, env, client)
	})
	t.Run("pipeline", func(t *testing.T) {
		scenarioPipeline(t, env, client)
	})
	t.Run("webhook", func(t *testing.T) {
		scenarioWebhook(t, env, client)
	})
	t.Run("billingcycle", func(t *testing.T) {
		scenarioBillingcycle(t, env, client)
	})
	t.Run("replay", func(t *testing.T) {
		scenarioReplay(t, env, client)
	})
	t.Run("idempotency", func(t *testing.T) {
		scenarioIdempotency(t, env, client)
	})
	t.Run("quota", func(t *testing.T) {
		scenarioQuota(t, env, client)
	})
}
