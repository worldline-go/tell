package tell

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectGRPC to use connection otel grpc endpoint, usually using to connect otel collector.
//
// grpc.WithBlock() disabled and it can connect later when collector exist.
func (c *Collector) ConnectGRPC(ctx context.Context, url string) error {
	conn, err := grpc.DialContext(ctx, url,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	c.Conn = conn

	return nil
}

// CloseGRPC to closing opened grpc connection.
func (c *Collector) CloseGRPC() error {
	if c.Conn == nil {
		return nil
	}

	if err := c.Conn.Close(); err != nil {
		return fmt.Errorf("failed to close gRPC connection: %w", err)
	}

	return nil
}
