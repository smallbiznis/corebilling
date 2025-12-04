package invoice

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/invoice/domain"
	invoicev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModuleGRPC registers the invoice service with the shared gRPC server.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

// RegisterGRPC attaches the invoice handler.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	invoicev1.RegisterInvoiceServiceServer(server, &grpcService{svc: svc})
}

type grpcService struct {
	invoicev1.UnimplementedInvoiceServiceServer
	svc *domain.Service
}

func (g *grpcService) GetInvoice(ctx context.Context, req *invoicev1.GetInvoiceRequest) (*invoicev1.GetInvoiceResponse, error) {
	inv, err := g.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &invoicev1.GetInvoiceResponse{Invoice: g.toProto(inv)}, nil
}

func (g *grpcService) ListInvoices(ctx context.Context, req *invoicev1.ListInvoicesRequest) (*invoicev1.ListInvoicesResponse, error) {
	items, err := g.svc.ListByTenant(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}

	resp := &invoicev1.ListInvoicesResponse{}
	for _, item := range items {
		resp.Invoices = append(resp.Invoices, g.toProto(item))
	}
	return resp, nil
}

func (g *grpcService) toProto(inv domain.Invoice) *invoicev1.Invoice {
	return &invoicev1.Invoice{
		Id:                 inv.ID,
		TenantId:           inv.TenantID,
		BillingPeriodStart: timestamppb.New(inv.BillingPeriodStart),
		BillingPeriodEnd:   timestamppb.New(inv.BillingPeriodEnd),
		TotalAmountCents:   inv.TotalCents,
		Status:             inv.Status,
		CreatedAt:          timestamppb.New(inv.CreatedAt),
	}
}
