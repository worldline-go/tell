package exporter

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc"
)

var defaultInterval = 2 * time.Second

type Otel struct {
	Conn *grpc.ClientConn
	OtelSetting
}

type OtelSetting struct {
	Interval time.Duration
}

// MetricCollector return a reader with using grpc connection.
func (m *Otel) Metric(ctx context.Context) (metricsdk.Reader, error) {
	if m.Interval == 0 {
		m.Interval = defaultInterval
	}
	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(m.Conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	reader := metricsdk.NewPeriodicReader(
		metricExporter, metricsdk.WithInterval(m.Interval),
	)

	return reader, nil
}
