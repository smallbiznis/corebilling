package idempotency

import (
	"os/exec"
	"testing"
)

func requireDocker(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skipf("docker not available: %v", err)
	}
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skipf("docker not available: %v", err)
	}
}
