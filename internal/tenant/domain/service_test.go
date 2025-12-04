package domain_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	tenantmock "github.com/smallbiznis/corebilling/internal/mocks/tenant"
	domain "github.com/smallbiznis/corebilling/internal/tenant/domain"
	tenantv1 "github.com/smallbiznis/go-genproto/smallbiznis/tenant/v1"
	"go.uber.org/zap"
)

func TestService_CreateTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := tenantmock.NewMockRepository(ctrl)
	var captured domain.Tenant
	mockRepo.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(domain.Tenant{})).DoAndReturn(func(_ context.Context, tnt domain.Tenant) error {
		captured = tnt
		return nil
	}).Times(1)

	service := domain.NewService(mockRepo, zap.NewNop())
	req := &tenantv1.CreateTenantRequest{
		Name: "acme",
		Slug: "acme-platform",
	}
	_, err := service.CreateTenant(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTenant returned error %v", err)
	}
	if captured.Slug != "acme-platform" {
		t.Fatalf("slug = %q, want acme-platform", captured.Slug)
	}
}

func TestService_UpdateTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	existing := domain.Tenant{
		ID:     "123",
		Name:   "old",
		Slug:   "old",
		Status: tenantv1.TenantStatus_TENANT_STATUS_ACTIVE,
	}
	repo := tenantmock.NewMockRepository(ctrl)
	repo.EXPECT().GetByID(gomock.Any(), "123").Return(existing, nil).Times(1)
	var updated domain.Tenant
	repo.EXPECT().Update(gomock.Any(), gomock.AssignableToTypeOf(domain.Tenant{})).DoAndReturn(func(_ context.Context, tnt domain.Tenant) error {
		updated = tnt
		return nil
	}).Times(1)

	service := domain.NewService(repo, zap.NewNop())
	req := &tenantv1.UpdateTenantRequest{
		Id:   "123",
		Name: "new-name",
		Slug: "new-slug",
	}
	_, err := service.UpdateTenant(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateTenant returned %v", err)
	}
	if updated.Name != "new-name" {
		t.Fatalf("updated name = %q, want new-name", updated.Name)
	}
	if updated.Slug != "new-slug" {
		t.Fatalf("updated slug = %q, want new-slug", updated.Slug)
	}
}
