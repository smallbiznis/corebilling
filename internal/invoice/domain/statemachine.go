package domain

import (
	"errors"
	"fmt"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	invoicev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// InvoiceLifecycle represents invoice-specific state transitions.
type InvoiceLifecycle string

const (
	InvoiceLifecycleCreated InvoiceLifecycle = "invoice.created"
	InvoiceLifecycleOpened  InvoiceLifecycle = "invoice.opened"
	InvoiceLifecyclePaid    InvoiceLifecycle = "invoice.paid"
	InvoiceLifecycleVoided  InvoiceLifecycle = "invoice.voided"
)

var invoiceTransitions = map[InvoiceLifecycle]transitionRuleInvoice{
	InvoiceLifecycleCreated: {from: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_UNSPECIFIED)}, to: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_DRAFT)}},
	InvoiceLifecycleOpened:  {from: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_DRAFT)}, to: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_OPEN)}},
	InvoiceLifecyclePaid:    {from: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_OPEN)}, to: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_PAID)}},
	InvoiceLifecycleVoided:  {from: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_DRAFT), invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_OPEN)}, to: []invoicev1.InvoiceStatus{invoiceStatus(invoicev1.InvoiceStatus_INVOICE_STATUS_VOID)}},
}

type transitionRuleInvoice struct {
	from []invoicev1.InvoiceStatus
	to   []invoicev1.InvoiceStatus
}

func invoiceStatus(val invoicev1.InvoiceStatus) invoicev1.InvoiceStatus {
	return val
}

var ErrInvalidInvoiceTransition = errors.New("invalid invoice transition")

// ApplyLifecycle applies a lifecycle event and emits a domain event when the status changes.
func (inv *Invoice) ApplyLifecycle(event InvoiceLifecycle, target invoicev1.InvoiceStatus) (*eventv1.Event, error) {
	rule, ok := invoiceTransitions[event]
	if !ok {
		return nil, fmt.Errorf("unknown invoice lifecycle %q", event)
	}
	current := invoicev1.InvoiceStatus(inv.Status)
	if !containsInvoiceStatus(rule.from, current) && current != invoicev1.InvoiceStatus_INVOICE_STATUS_UNSPECIFIED {
		return nil, fmt.Errorf("%w: %s -> %s", ErrInvalidInvoiceTransition, current.String(), target.String())
	}
	if !containsInvoiceStatus(rule.to, target) {
		return nil, fmt.Errorf("%w: invalid target %s for %s", ErrInvalidInvoiceTransition, target.String(), event)
	}
	inv.Status = int32(target)
	return buildInvoiceEvent(inv, target.String())
}

func containsInvoiceStatus(list []invoicev1.InvoiceStatus, status invoicev1.InvoiceStatus) bool {
	for _, s := range list {
		if s == status {
			return true
		}
	}
	return false
}

func buildInvoiceEvent(inv *Invoice, status string) (*eventv1.Event, error) {
	payload, err := structpb.NewStruct(map[string]interface{}{
		"invoice_id": inv.ID,
		"status":     status,
	})
	if err != nil {
		return nil, err
	}
	return &eventv1.Event{
		Subject:  "invoice.status.changed",
		TenantId: inv.TenantID,
		Data:     payload,
	}, nil
}
