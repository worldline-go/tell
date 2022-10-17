package tell

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Collector hold metric and trace informations.
type Collector struct {
	Conn *grpc.ClientConn
}

// ConnectGRPC to use connection otel grpc endpoint, usually using to connect otel collector.
func (c *Collector) ConnectGRPC(ctx context.Context, url string) (*Collector, error) {
	conn, err := grpc.DialContext(ctx, url,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return c, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	c.Conn = conn

	return c, nil
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
