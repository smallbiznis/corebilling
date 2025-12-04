package subscription

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/smallbiznis/corebilling/internal/subscription/domain"
	subscriptionv1 "github.com/smallbiznis/go-genproto/smallbiznis/subscription/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func setupService() (*grpcService, *domain.TestRepository) {
	repo := domain.NewTestRepository()
	svc := domain.NewService(repo, zap.NewNop())
	return &grpcService{svc: svc}, repo
}

func TestCreateSubscriptionValidation(t *testing.T) {
	srv, _ := setupService()
	_, err := srv.CreateSubscription(context.Background(), &subscriptionv1.CreateSubscriptionRequest{})
	if err == nil || status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument error, got %v", err)
	}
}

func TestCreateSubscriptionSuccess(t *testing.T) {
	srv, repo := setupService()
	req := &subscriptionv1.CreateSubscriptionRequest{
		TenantId:     "tenant",
		CustomerId:   "cust",
		PriceId:      "price",
		TrialStartAt: timestamppb.New(time.Now().Add(-time.Hour)),
		TrialEndAt:   timestamppb.New(time.Now().Add(time.Hour)),
	}
	resp, err := srv.CreateSubscription(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Id == "" {
		t.Fatal("expected id")
	}
	if _, ok := repo.Subs[resp.Id]; !ok {
		t.Fatal("subscription not stored")
	}
}

func TestListSubscriptionsPagination(t *testing.T) {
	srv, repo := setupService()
	for i := 0; i < 3; i++ {
		sub := domain.Subscription{
			ID:         "sub" + strconv.Itoa(i),
			TenantID:   "tenant",
			CustomerID: "cust",
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Minute),
		}
		if err := repo.Create(context.Background(), sub); err != nil {
			t.Fatal(err)
		}
	}
	resp, err := srv.ListSubscriptions(context.Background(), &subscriptionv1.ListSubscriptionsRequest{
		TenantId: "tenant",
		PageSize: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.NextPageToken == "" {
		t.Fatal("expected pagination token")
	}
	if len(resp.Subscriptions) != 1 {
		t.Fatalf("expected 1 record, got %d", len(resp.Subscriptions))
	}
}

func TestCancelSubscription(t *testing.T) {
	srv, repo := setupService()
	sub := domain.Subscription{
		ID:                 "sub-1",
		TenantID:           "tenant",
		CustomerID:         "cust",
		PriceID:            "price",
		CurrentPeriodEnd:   time.Now().Add(24 * time.Hour),
		CurrentPeriodStart: time.Now(),
		Status:             int32(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE),
		AutoRenew:          true,
	}
	if err := repo.Create(context.Background(), sub); err != nil {
		t.Fatal(err)
	}
	_, err := srv.CancelSubscription(context.Background(), &subscriptionv1.CancelSubscriptionRequest{
		Id:                sub.ID,
		TenantId:          sub.TenantID,
		CancelAtPeriodEnd: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := repo.GetByID(context.Background(), sub.ID)
	if stored.Status != int32(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED) {
		t.Fatalf("expected canceled status, got %d", stored.Status)
	}
	if stored.CanceledAt != nil {
		t.Fatal("expected canceled_at nil when canceling at period end")
	}
	if stored.CancelAt == nil {
		t.Fatal("expected cancel_at to be set at period end")
	}

	_, err = srv.CancelSubscription(context.Background(), &subscriptionv1.CancelSubscriptionRequest{
		Id:                sub.ID,
		TenantId:          sub.TenantID,
		CancelAtPeriodEnd: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, _ = repo.GetByID(context.Background(), sub.ID)
	if stored.CanceledAt == nil {
		t.Fatal("expected canceled_at set for immediate cancel")
	}
}
