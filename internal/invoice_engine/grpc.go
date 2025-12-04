package invoice_engine

import (
	"github.com/smallbiznis/corebilling/internal/invoice_engine/domain"
	invoiceenginev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice_engine/v1"
	"google.golang.org/grpc"
)

// RegisterGRPC attaches the invoice engine service to the shared server.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	invoiceenginev1.RegisterInvoiceEngineServiceServer(server, svc)
}
