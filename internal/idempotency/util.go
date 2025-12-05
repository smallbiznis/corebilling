package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
)

// ComputeHash returns the SHA256 hash of the provided body.
func ComputeHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
