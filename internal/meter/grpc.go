package meter

import (
	"github.com/smallbiznis/corebilling/internal/meter/domain"
	meterv1 "github.com/smallbiznis/go-genproto/smallbiznis/meter/v1"
	"google.golang.org/grpc"
)

func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	meterv1.RegisterMeterServiceServer(server, svc)
}
