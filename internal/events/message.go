package events

import (
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
)

// Event represents the canonical event envelope used across providers.
type Event = eventv1.Event
