package domain

import (
	"context"
	"strconv"
	"time"

	auditv1 "github.com/smallbiznis/go-genproto/smallbiznis/audit/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultAuditPageSize = 50

// Service implements auditv1.AuditServiceServer.
type Service struct {
	auditv1.UnimplementedAuditServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService constructs a new audit service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("audit.service")}
}

func (s *Service) CreateAuditEvent(ctx context.Context, req *auditv1.CreateAuditEventRequest) (*auditv1.CreateAuditEventResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetAction() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and action are required")
	}

	event := AuditEvent{
		ID:           strconv.FormatInt(time.Now().UnixNano(), 10),
		TenantID:     req.GetTenantId(),
		ActorType:    req.GetActorType(),
		ActorID:      req.GetActorId(),
		Action:       req.GetAction(),
		ActionType:   req.GetActionType(),
		ResourceType: req.GetResourceType(),
		ResourceID:   req.GetResourceId(),
		OldValues:    structToMap(req.GetOldValues()),
		NewValues:    structToMap(req.GetNewValues()),
		IpAddress:    req.GetIpAddress(),
		UserAgent:    req.GetUserAgent(),
		Metadata:     structToMap(req.GetMetadata()),
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, event); err != nil {
		s.logger.Error("persisting audit event failed", zap.Error(err))
		return nil, err
	}

	return &auditv1.CreateAuditEventResponse{Event: s.toProto(event)}, nil
}

func (s *Service) ListAuditEvents(ctx context.Context, req *auditv1.ListAuditEventsRequest) (*auditv1.ListAuditEventsResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultAuditPageSize
	}
	offset := parsePageToken(req.GetPageToken())
	filter := ListFilter{
		TenantID:     req.GetTenantId(),
		ResourceType: req.GetResourceType(),
		ResourceID:   req.GetResourceId(),
		ActorID:      req.GetActorId(),
		ActionType:   req.GetActionType(),
		Limit:        pageSize,
		Offset:       offset,
	}

	events, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	resp := &auditv1.ListAuditEventsResponse{}
	for _, event := range events {
		resp.Events = append(resp.Events, s.toProto(event))
	}
	if len(events) == pageSize {
		resp.NextPageToken = strconv.Itoa(offset + len(events))
	}
	return resp, nil
}

func (s *Service) GetAuditEvent(ctx context.Context, req *auditv1.GetAuditEventRequest) (*auditv1.GetAuditEventResponse, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	event, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &auditv1.GetAuditEventResponse{Event: s.toProto(event)}, nil
}

func (s *Service) toProto(event AuditEvent) *auditv1.AuditEvent {
	return &auditv1.AuditEvent{
		Id:           event.ID,
		TenantId:     event.TenantID,
		ActorType:    event.ActorType,
		ActorId:      event.ActorID,
		Action:       event.Action,
		ActionType:   event.ActionType,
		ResourceType: event.ResourceType,
		ResourceId:   event.ResourceID,
		OldValues:    mapToStruct(event.OldValues),
		NewValues:    mapToStruct(event.NewValues),
		IpAddress:    event.IpAddress,
		UserAgent:    event.UserAgent,
		Metadata:     mapToStruct(event.Metadata),
		CreatedAt:    timestamppb.New(event.CreatedAt),
	}
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
