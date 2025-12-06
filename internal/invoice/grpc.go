package invoice

import (
	"context"
	"strconv"

	"github.com/smallbiznis/corebilling/internal/invoice/domain"
	invoicev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModuleGRPC registers the invoice service with the shared gRPC server.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

func RegisterService(svc *domain.Service) *grpcService {
	return &grpcService{svc: svc}
}

// RegisterGRPC attaches the invoice handler.
func RegisterGRPC(server *grpc.Server, svc *grpcService) {
	invoicev1.RegisterInvoiceServiceServer(server, svc)
}

type grpcService struct {
	invoicev1.UnimplementedInvoiceServiceServer
	svc *domain.Service
}

const (
	defaultInvoicePageSize = 50
	maxInvoicePageSize     = 200
)

func (g *grpcService) GetInvoice(ctx context.Context, req *invoicev1.GetInvoiceRequest) (*invoicev1.Invoice, error) {
	if req == nil || req.GetId() == "" || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and invoice id required")
	}
	inv, err := g.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if inv.TenantID != req.GetTenantId() {
		return nil, status.Error(codes.NotFound, "invoice not found")
	}
	return g.toProto(inv), nil
}

func (g *grpcService) ListInvoices(ctx context.Context, req *invoicev1.ListInvoicesRequest) (*invoicev1.ListInvoicesResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id required")
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultInvoicePageSize
	}
	if pageSize > maxInvoicePageSize {
		pageSize = maxInvoicePageSize
	}

	offset := parsePageToken(req.GetPageToken())

	items, hasMore, err := g.svc.List(ctx, domain.ListInvoicesFilter{
		TenantID:       req.GetTenantId(),
		CustomerID:     req.GetCustomerId(),
		SubscriptionID: req.GetSubscriptionId(),
		Status:         int32(req.GetStatus()),
		Limit:          pageSize,
		Offset:         offset,
	})
	if err != nil {
		return nil, err
	}

	resp := &invoicev1.ListInvoicesResponse{}
	for _, item := range items {
		resp.Invoices = append(resp.Invoices, g.toProto(item))
	}
	if hasMore {
		resp.NextPageToken = strconv.Itoa(offset + len(items))
	}
	return resp, nil
}

func (g *grpcService) toProto(inv domain.Invoice) *invoicev1.Invoice {
	var issuedAt, dueAt, paidAt *timestamppb.Timestamp
	if inv.IssuedAt != nil {
		issuedAt = timestamppb.New(*inv.IssuedAt)
	}
	if inv.DueAt != nil {
		dueAt = timestamppb.New(*inv.DueAt)
	}
	if inv.PaidAt != nil {
		paidAt = timestamppb.New(*inv.PaidAt)
	}

	return &invoicev1.Invoice{
		Id:             inv.ID,
		TenantId:       inv.TenantID,
		CustomerId:     inv.CustomerID,
		SubscriptionId: inv.SubscriptionID,
		Status:         invoicev1.InvoiceStatus(inv.Status),
		CurrencyCode:   inv.CurrencyCode,
		TotalCents:     inv.TotalCents,
		SubtotalCents:  inv.SubtotalCents,
		TaxCents:       inv.TaxCents,
		InvoiceNumber:  inv.InvoiceNumber,
		IssuedAt:       issuedAt,
		DueAt:          dueAt,
		PaidAt:         paidAt,
		Metadata:       mapToStruct(inv.Metadata),
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
