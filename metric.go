package tell

import (
	"go.opentelemetry.io/otel/metric/global"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/view"
	"go.opentelemetry.io/otel/sdk/resource"
)

// MetricProvider adds required labels to readers and return a meterprovider.
// Run shutdown will flush any remaining spans and shut down the exporter.
//
// MetricProvider set the provider to collector and return it.
func (c *Collector) MetricProvider(views []view.View, mReaders ...metricsdk.Reader) *Collector {
	// Set resource for auto show some attributes about this service
	// Set OTEL_SERVICE_NAME or OTEL_RESOURCE_ATTRIBUTES

	options := []metricsdk.Option{metricsdk.WithResource(resource.Environment())}

	for _, mReader := range mReaders {
		options = append(options, metricsdk.WithReader(mReader, views...))
	}

	meterProvider := metricsdk.NewMeterProvider(
		options...,
	)

	c.MeterProvider = meterProvider

	return c
}

// SetMetricProviderGlobal to globally set provider.
func (c *Collector) SetMetricProviderGlobal() *Collector {
	global.SetMeterProvider(c.MeterProvider)

	return c
}
