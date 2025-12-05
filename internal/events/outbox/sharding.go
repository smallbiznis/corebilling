package outbox

import (
	"crypto/md5"
	"encoding/binary"
)

// ShardIndex deterministically maps an event ID to a shard bucket.
func ShardIndex(id string, shardTotal int) int {
	if shardTotal <= 1 {
		return 0
	}
	hash := md5.Sum([]byte(id))
	val := binary.BigEndian.Uint64(hash[:8])
	return int(val % uint64(shardTotal))
}
