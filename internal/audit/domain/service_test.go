package domain_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	domain "github.com/smallbiznis/corebilling/internal/audit/domain"
	auditmock "github.com/smallbiznis/corebilling/internal/mocks/audit"
	auditv1 "github.com/smallbiznis/go-genproto/smallbiznis/audit/v1"
	"go.uber.org/zap"
)

func TestService_CreateAuditEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := auditmock.NewMockRepository(ctrl)
	var captured domain.AuditEvent
	mockRepo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(domain.AuditEvent{})).Do(func(_ context.Context, ev domain.AuditEvent) {
		captured = ev
	}).Return(nil).Times(1)

	service := domain.NewService(mockRepo, zap.NewNop())
	req := &auditv1.CreateAuditEventRequest{
		TenantId: "tenant",
		Action:   "created",
	}
	_, err := service.CreateAuditEvent(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := captured.TenantID; got != "tenant" {
		t.Fatalf("tenant id = %q, want tenant", got)
	}
}

func TestService_ListAuditEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := auditmock.NewMockRepository(ctrl)
	var capturedFilter domain.ListFilter
	mockRepo.EXPECT().List(gomock.Any(), gomock.AssignableToTypeOf(domain.ListFilter{})).DoAndReturn(func(_ context.Context, f domain.ListFilter) ([]domain.AuditEvent, error) {
		capturedFilter = f
		return []domain.AuditEvent{{ID: "1"}}, nil
	}).Times(1)

	service := domain.NewService(mockRepo, zap.NewNop())
	resp, err := service.ListAuditEvents(context.Background(), &auditv1.ListAuditEventsRequest{
		TenantId: "tenant",
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("ListAuditEvents returned err %v", err)
	}
	if resp.NextPageToken != "1" {
		t.Fatalf("next page token = %q, want %q", resp.NextPageToken, "1")
	}
	if capturedFilter.Limit != 1 {
		t.Fatalf("list limit = %d, want %d", capturedFilter.Limit, 1)
	}
}
