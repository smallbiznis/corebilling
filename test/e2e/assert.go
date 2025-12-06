package e2e

import (
	"testing"
)

func mustStatus(t *testing.T, want, got string) {
	t.Helper()
	if want != got {
		t.Fatalf("unexpected status: want=%s got=%s", want, got)
	}
}

func mustJSONEqual(t *testing.T, a, b []byte) {
	t.Helper()
	if string(a) != string(b) {
		t.Fatalf("json mismatch\nwant: %s\ngot: %s", string(a), string(b))
	}
}
