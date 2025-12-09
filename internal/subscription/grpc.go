package subscription

import (
	"context"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/smallbiznis/corebilling/internal/subscription/domain"
	subscriptionv1 "github.com/smallbiznis/go-genproto/smallbiznis/subscription/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModuleGRPC registers the subscription service.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

func RegisterService(svc *domain.Service, genID *snowflake.Node) *grpcService {
	return NewGrpcService(svc, genID)
}

// RegisterGRPC attaches the subscription handler.
func RegisterGRPC(server *grpc.Server, svc *grpcService) {
	subscriptionv1.RegisterSubscriptionServiceServer(server, svc)
}

type grpcService struct {
	subscriptionv1.UnimplementedSubscriptionServiceServer
	svc *domain.Service

	genID *snowflake.Node
}

const (
	defaultSubscriptionPageSize = 50
	maxSubscriptionPageSize     = 200
)

func NewGrpcService(svc *domain.Service, genID *snowflake.Node) *grpcService {
	return &grpcService{svc: svc, genID: genID}
}

func (g *grpcService) CreateSubscription(ctx context.Context, req *subscriptionv1.CreateSubscriptionRequest) (*subscriptionv1.Subscription, error) {
	if req == nil || req.GetTenantId() == "" || req.GetCustomerId() == "" || req.GetPriceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id, customer_id and price_id required")
	}

	now := time.Now().UTC()
	startAt := now
	if trial := req.GetTrialStartAt(); trial != nil {
		startAt = trial.AsTime()
	}
	currentPeriodStart := startAt
	currentPeriodEnd := startAt.AddDate(0, 1, 0)

	status := subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_ACTIVE
	if req.GetTrialStartAt() != nil && req.GetTrialEndAt() != nil {
		status = subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_TRIALING
	}

	sub := domain.Subscription{
		ID:                 g.genID.Generate().String(),
		TenantID:           req.GetTenantId(),
		CustomerID:         req.GetCustomerId(),
		PriceID:            req.GetPriceId(),
		Status:             int32(status),
		AutoRenew:          req.GetAutoRenew(),
		StartAt:            startAt,
		CurrentPeriodStart: currentPeriodStart,
		CurrentPeriodEnd:   currentPeriodEnd,
		TrialStartAt:       toTimePtr(req.GetTrialStartAt()),
		TrialEndAt:         toTimePtr(req.GetTrialEndAt()),
		Metadata:           structToMap(req.GetMetadata()),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := g.svc.Create(ctx, sub); err != nil {
		return nil, err
	}
	return g.toProto(sub), nil
}

func (g *grpcService) GetSubscription(ctx context.Context, req *subscriptionv1.GetSubscriptionRequest) (*subscriptionv1.Subscription, error) {
	if req == nil || req.GetId() == "" || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and subscription id required")
	}
	sub, err := g.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if sub.TenantID != req.GetTenantId() {
		return nil, status.Error(codes.NotFound, "subscription not found")
	}
	return g.toProto(sub), nil
}

func (g *grpcService) ListSubscriptions(ctx context.Context, req *subscriptionv1.ListSubscriptionsRequest) (*subscriptionv1.ListSubscriptionsResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id required")
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultSubscriptionPageSize
	}
	if pageSize > maxSubscriptionPageSize {
		pageSize = maxSubscriptionPageSize
	}

	offset := parsePageToken(req.GetPageToken())

	subs, hasMore, err := g.svc.List(ctx, domain.ListSubscriptionsFilter{
		TenantID:   req.GetTenantId(),
		CustomerID: req.GetCustomerId(),
		Limit:      pageSize,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	resp := &subscriptionv1.ListSubscriptionsResponse{}
	for _, sub := range subs {
		resp.Subscriptions = append(resp.Subscriptions, g.toProto(sub))
	}
	if hasMore {
		resp.NextPageToken = strconv.Itoa(offset + len(subs))
	}
	return resp, nil
}

func (g *grpcService) CancelSubscription(ctx context.Context, req *subscriptionv1.CancelSubscriptionRequest) (*subscriptionv1.Subscription, error) {
	if req == nil || req.GetId() == "" || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and subscription id required")
	}
	sub, err := g.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if sub.TenantID != req.GetTenantId() {
		return nil, status.Error(codes.NotFound, "subscription not found")
	}

	now := time.Now().UTC()
	sub.AutoRenew = false
	sub.UpdatedAt = now
	sub.Status = int32(subscriptionv1.SubscriptionStatus_SUBSCRIPTION_STATUS_CANCELED)
	if req.GetCancelAtPeriodEnd() {
		cancelAt := sub.CurrentPeriodEnd
		sub.CancelAt = &cancelAt
		sub.CanceledAt = nil
	} else {
		cancelAt := now
		sub.CancelAt = &cancelAt
		sub.CanceledAt = &cancelAt
	}

	if err := g.svc.Update(ctx, sub); err != nil {
		return nil, err
	}
	return g.toProto(sub), nil
}

func (g *grpcService) toProto(sub domain.Subscription) *subscriptionv1.Subscription {
	var issuedAt, periodStart, periodEnd, trialStart, trialEnd, cancelAt, canceledAt *timestamppb.Timestamp
	if !sub.StartAt.IsZero() {
		issuedAt = timestamppb.New(sub.StartAt)
	}
	if !sub.CurrentPeriodStart.IsZero() {
		periodStart = timestamppb.New(sub.CurrentPeriodStart)
	}
	if !sub.CurrentPeriodEnd.IsZero() {
		periodEnd = timestamppb.New(sub.CurrentPeriodEnd)
	}
	if sub.TrialStartAt != nil {
		trialStart = timestamppb.New(*sub.TrialStartAt)
	}
	if sub.TrialEndAt != nil {
		trialEnd = timestamppb.New(*sub.TrialEndAt)
	}
	if sub.CancelAt != nil {
		cancelAt = timestamppb.New(*sub.CancelAt)
	}
	if sub.CanceledAt != nil {
		canceledAt = timestamppb.New(*sub.CanceledAt)
	}

	return &subscriptionv1.Subscription{
		Id:                 sub.ID,
		TenantId:           sub.TenantID,
		CustomerId:         sub.CustomerID,
		PriceId:            sub.PriceID,
		Status:             subscriptionv1.SubscriptionStatus(sub.Status),
		AutoRenew:          sub.AutoRenew,
		StartAt:            issuedAt,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		TrialStartAt:       trialStart,
		TrialEndAt:         trialEnd,
		CancelAt:           cancelAt,
		CanceledAt:         canceledAt,
		Metadata:           mapToStruct(sub.Metadata),
	}
}

func parsePageToken(token string) int {
	if token == "" {
		return 0
	}
	val, err := strconv.Atoi(token)
	if err != nil || val < 0 {
		return 0
	}
	return val
}

func toTimePtr(ts *timestamppb.Timestamp) *time.Time {
	if ts == nil {
		return nil
	}
	t := ts.AsTime()
	return &t
}

func structToMap(value *structpb.Struct) map[string]interface{} {
	if value == nil {
		return nil
	}
	return value.AsMap()
}

func mapToStruct(value map[string]interface{}) *structpb.Struct {
	if len(value) == 0 {
		return nil
	}
	s, err := structpb.NewStruct(value)
	if err != nil {
		return nil
	}
	return s
}
