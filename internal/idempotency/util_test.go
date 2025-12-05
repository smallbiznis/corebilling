package idempotency

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeHashDeterministic(t *testing.T) {
	body := []byte("example body")
	hash1 := ComputeHash(body)
	hash2 := ComputeHash(body)

	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}

func TestComputeHashDifferentBodies(t *testing.T) {
	hash1 := ComputeHash([]byte("first"))
	hash2 := ComputeHash([]byte("second"))

	assert.NotEqual(t, hash1, hash2)
}

func TestComputeHashEmpty(t *testing.T) {
	hash := ComputeHash(nil)

	assert.NotEmpty(t, hash)
	assert.Equal(t, hash, ComputeHash([]byte{}))
}
