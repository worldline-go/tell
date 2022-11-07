package tell

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/tglobal"
)

var defaultInterval = 2 * time.Second

// MetricProvider adds required labels to readers and return a meterprovider.
// Run shutdown will flush any remaining spans and shut down the exporter.
//
// MetricProvider set the provider to collector and return it.
func (c *Collector) MetricProvider(ctx context.Context, cfg MetricProviderSettings) error {
	if c.Conn == nil {
		return ErrSetConnetion
	}

	interval := cfg.Interval
	if interval == 0 {
		interval = defaultInterval
	}

	// Set up a trace exporter
	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(c.Conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	c.MetricReader = metricsdk.NewPeriodicReader(
		metricExporter, metricsdk.WithInterval(interval),
	)

	// Set resource for auto show some attributes about this service
	// Set OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES

	options := []metricsdk.Option{
		metricsdk.WithResource(resource.Environment()),
		metricsdk.WithReader(c.MetricReader, tglobal.MetricViews.GetViews()...),
	}

	meterProvider := metricsdk.NewMeterProvider(
		options...,
	)

	c.MeterProviderSDK = meterProvider
	c.MeterProvider = meterProvider

	return nil
}

// SetMetricProviderGlobal to globally set provider.
func (c *Collector) SetMetricProviderGlobal() *Collector {
	global.SetMeterProvider(c.MeterProvider)

	return c
}
