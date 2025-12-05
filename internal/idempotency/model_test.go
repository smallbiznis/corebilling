package idempotency

import "testing"

func TestStatusStringValues(t *testing.T) {
	if StatusProcessing != "PROCESSING" {
		t.Fatalf("expected StatusProcessing to be PROCESSING, got %s", StatusProcessing)
	}
	if StatusCompleted != "COMPLETED" {
		t.Fatalf("expected StatusCompleted to be COMPLETED, got %s", StatusCompleted)
	}
}
