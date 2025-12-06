package e2e

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// apiClient wraps shared gRPC connection and default metadata.
type apiClient struct {
	conn *grpc.ClientConn
}

func newAPIClient(ctx context.Context, addr string) (*apiClient, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &apiClient{conn: conn}, nil
}

func (c *apiClient) close() {
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// withTenant attaches required headers.
func withTenant(ctx context.Context, tenant, apiKey, idem string) context.Context {
	md := metadata.MD{}
	if tenant != "" {
		md.Append("x-tenant-id", tenant)
	}
	if apiKey != "" {
		md.Append("x-api-key", apiKey)
	}
	if idem != "" {
		md.Append("x-idempotency-key", idem)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// newTimeoutCtx returns a new context with default timeout.
func newTimeoutCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}
