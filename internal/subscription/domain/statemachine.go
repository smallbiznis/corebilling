package domain

import (
	"errors"
	"fmt"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	subscriptionv1 "github.com/smallbiznis/go-genproto/smallbiznis/subscription/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// SubscriptionLifecycle represents lifecycle events for a subscription state machine.
type SubscriptionLifecycle string

const (
	SubscriptionLifecycleCreated      SubscriptionLifecycle = "subscription.created"
	SubscriptionLifecycleTrialStarted SubscriptionLifecycle = "subscription.trial_started"
	SubscriptionLifecycleActivated    SubscriptionLifecycle = "subscription.activated"
	SubscriptionLifecycleCanceled     SubscriptionLifecycle = "subscription.canceled"
)

var subscriptionTransitions = map[SubscriptionLifecycle]transitionRule{
	SubscriptionLifecycleCreated:      {from: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED)}, to: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING), subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE)}},
	SubscriptionLifecycleTrialStarted: {from: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING)}, to: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING)}},
	SubscriptionLifecycleActivated:    {from: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING), subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE)}, to: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE)}},
	SubscriptionLifecycleCanceled:     {from: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE), subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING)}, to: []subscriptionv1.SubscriptionStatus{subsStatus(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED)}},
}

type transitionRule struct {
	from []subscriptionv1.SubscriptionStatus
	to   []subscriptionv1.SubscriptionStatus
}

func containsStatus(list []subscriptionv1.SubscriptionStatus, status subscriptionv1.SubscriptionStatus) bool {
	for _, s := range list {
		if s == status {
			return true
		}
	}
	return false
}

func subsStatus(val subscriptionv1.SubscriptionStatus) subscriptionv1.SubscriptionStatus {
	return val
}

var ErrInvalidSubscriptionTransition = errors.New("invalid subscription transition")

// ApplyLifecycle updates the subscription state based on the provided event.
func (s *Subscription) ApplyLifecycle(event SubscriptionLifecycle, target subscriptionv1.SubscriptionStatus) (*eventv1.Event, error) {
	rule, ok := subscriptionTransitions[event]
	if !ok {
		return nil, fmt.Errorf("unknown subscription lifecycle %q", event)
	}
	current := subscriptionv1.SubscriptionStatus(s.Status)
	if !containsStatus(rule.from, current) && current != subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_UNSPECIFIED {
		return nil, fmt.Errorf("%w: %s -> %s", ErrInvalidSubscriptionTransition, current.String(), target.String())
	}
	if !containsStatus(rule.to, target) {
		return nil, fmt.Errorf("%w: invalid target %s for %s", ErrInvalidSubscriptionTransition, target.String(), event)
	}
	s.Status = int32(target)
	return buildSubscriptionEvent(s, target.String())
}

func buildSubscriptionEvent(sub *Subscription, status string) (*eventv1.Event, error) {
	payload, err := structpb.NewStruct(map[string]interface{}{
		"subscription_id": sub.ID,
		"status":          status,
	})
	if err != nil {
		return nil, err
	}
	return &eventv1.Event{
		Subject:  "subscription.status.changed",
		TenantId: sub.TenantID,
		Data:     payload,
	}, nil
}
