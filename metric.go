package tell

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/view"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MetricCollector return a reader with using grpc connection.
func (c *Collector) MetricCollector(ctx context.Context) (metricsdk.Reader, error) {
	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(c.Conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	reader := metricsdk.NewPeriodicReader(
		metricExporter, metricsdk.WithInterval(2*time.Second),
	)

	return reader, nil
}

// MetricProvider adds required labels to readers and return a meterprovider.
// Run shutdown will flush any remaining spans and shut down the exporter.
func (c *Collector) MetricProvider(views []view.View, mReaders ...metricsdk.Reader) *metricsdk.MeterProvider {
	// Set resource for auto show some attributes about this service
	// you can use resource.Default()
	// Set OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES
	options := []metricsdk.Option{metricsdk.WithResource(resource.Default())}

	for _, mReader := range mReaders {
		options = append(options, metricsdk.WithReader(mReader, views...))
	}

	meterProvider := metricsdk.NewMeterProvider(
		options...,
	)

	return meterProvider
}

func (c *Collector) SetMetricProviderGlobal(mp *metricsdk.MeterProvider) {
	global.SetMeterProvider(mp)
}
