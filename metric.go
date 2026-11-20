package tell

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

var defaultInterval = 5 * time.Second

const envMetricExportInterval = "OTEL_METRIC_EXPORT_INTERVAL"

// MetricProvider adds required labels to readers and return a meterprovider.
// Run shutdown will flush any remaining spans and shut down the exporter.
//
// MetricProvider set the provider to collector and return it.
func (c *Collector) MetricProvider(ctx context.Context, cfg MetricProviderSettings) error {
	// Set up a metric exporter
	metricExporter, err := c.metricExporter(ctx)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Priority: config value, then OTEL_METRIC_EXPORT_INTERVAL (read by the SDK), then defaultInterval.
	// WithInterval overrides the env value, so it is only set when the env is not given.
	var readerOpts []metricsdk.PeriodicReaderOption
	switch {
	case cfg.Interval > 0:
		readerOpts = append(readerOpts, metricsdk.WithInterval(cfg.Interval))
	case !validEnvMillis(envMetricExportInterval):
		readerOpts = append(readerOpts, metricsdk.WithInterval(defaultInterval))
	}

	c.MetricReader = metricsdk.NewPeriodicReader(metricExporter, readerOpts...)

	options := []metricsdk.Option{
		metricsdk.WithResource(c.Resource()),
		metricsdk.WithView(cfg.Views...),
		metricsdk.WithReader(c.MetricReader),
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
	otel.SetMeterProvider(c.MeterProvider)

	return c
}
