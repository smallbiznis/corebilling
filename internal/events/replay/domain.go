package replay

import "time"

// ReplayOptions define filters and toggles for replay execution.
type ReplayOptions struct {
	TenantID  *string
	EventType *string
	Since     *time.Time
	Until     *time.Time
	DryRun    bool // do not publish events, only log actions
}

const (
	// ReplayFlagKey marks replayed events in metadata.
	ReplayFlagKey      = "replay"
	ReplayBatchKey     = "replay_batch"
	ReplaySourceKey    = "replay_source"
	ReplayTimestampKey = "replay_at"
)
