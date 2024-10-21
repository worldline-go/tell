package tell

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ConnectGRPC to use connection otel grpc endpoint, usually using to connect otel collector.
func (c *Collector) ConnectGRPC(_ context.Context, url string, opts ...grpc.DialOption) error {
	// grpc.WithBlock() disabled and it can connect later when collector exist.
	opts = append([]grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, opts...)

	conn, err := grpc.NewClient(url, opts...)
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
