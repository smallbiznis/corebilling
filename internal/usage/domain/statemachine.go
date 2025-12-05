package domain

import (
	"errors"
	"fmt"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// UsageLifecycle governs transitions within usage ingestion.
type UsageLifecycle string

const (
	UsageLifecycleReported UsageLifecycle = "usage.reported"
	UsageLifecycleRated    UsageLifecycle = "usage.rated"
	UsageLifecycleBilled   UsageLifecycle = "usage.billed"
)

type usageState string

const (
	UsageStateUnknown  usageState = ""
	UsageStateReported usageState = "reported"
	UsageStateRated    usageState = "rated"
	UsageStateBilled   usageState = "billed"
)

var usageTransitions = map[UsageLifecycle]transitionUsage{
	UsageLifecycleReported: {from: []usageState{UsageStateUnknown}, to: UsageStateReported},
	UsageLifecycleRated:    {from: []usageState{UsageStateReported}, to: UsageStateRated},
	UsageLifecycleBilled:   {from: []usageState{UsageStateRated}, to: UsageStateBilled},
}

type transitionUsage struct {
	from []usageState
	to   usageState
}

var ErrInvalidUsageTransition = errors.New("invalid usage transition")

// ApplyLifecycle mutates the usage record's state according to the lifecycle.
func (u *UsageRecord) ApplyLifecycle(event UsageLifecycle) (*eventv1.Event, error) {
	rule, ok := usageTransitions[event]
	if !ok {
		return nil, fmt.Errorf("unknown usage lifecycle %q", event)
	}
	current := usageState(u.State)
	if !containsUsageState(rule.from, current) {
		return nil, fmt.Errorf("%w: %s -> %s", ErrInvalidUsageTransition, current, rule.to)
	}
	u.State = string(rule.to)
	return buildUsageEvent(u, rule.to)
}

func containsUsageState(list []usageState, state usageState) bool {
	for _, s := range list {
		if s == state {
			return true
		}
	}
	return false
}

func buildUsageEvent(u *UsageRecord, state usageState) (*eventv1.Event, error) {
	payload, err := structpb.NewStruct(map[string]interface{}{
		"usage_id": u.ID,
		"state":    string(state),
	})
	if err != nil {
		return nil, err
	}
	return &eventv1.Event{
		Subject:  "usage.status.changed",
		TenantId: u.TenantID,
		Data:     payload,
	}, nil
}

func subset(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
