package tell

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/view"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/config"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/metric/exporter"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/metric/instrumentation/metricecho"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/types"
)

type Config = config.Config

const defaultShutdownTimeOut = 2 * time.Second

// Collector hold metric and trace informations.
type Collector struct {
	Conn *grpc.ClientConn
	// Attributes have common attributes.
	Attributes map[string]interface{}
	// metrics
	MeterProvider *metricsdk.MeterProvider
	MetricReaders MetricReaders
	MetricViews   []view.View
	// traces
	TracerProvider *tracesdk.TracerProvider
	// ShutdownTimeOut for closing providers, default 2 seconds.
	ShutdownTimeOut time.Duration
}

type MetricReaders struct {
	Otel       metricsdk.Reader
	Prometheus *prometheus.Exporter
}

// New generate collectors based on configuration.
func New(ctx context.Context, cfg Config) (*Collector, error) {
	c := new(Collector)
	c.Attributes = cfg.Attributes

	// check grpc need
	if cfg.IsGRPC() {
		if err := c.ConnectGRPC(ctx, cfg.Collector); err != nil {
			return nil, err
		}
	}

	// if enabled additional views add here
	var views []view.View

	// set views
	for _, v := range cfg.GetEnabledViews() {
		switch v {
		case types.ViewRequestDuration:
			metricEchoViews, err := metricecho.GetViews()
			if err != nil {
				return nil, fmt.Errorf("failed to get metricecho views; %w", err)
			}

			views = append(views, metricEchoViews...)
		}
	}

	metricsEnabled := cfg.GetEnabledMetrics()
	if len(metricsEnabled) > 0 {
		// add meter provider for generate general metric provider
		var readers []metricsdk.Reader
		// set metrics
		for _, v := range metricsEnabled {
			switch v {
			case types.MetricOtel:
				otelExp := exporter.Otel{Conn: c.Conn, OtelSetting: cfg.MetricsSettings.Otel}

				otelReader, err := otelExp.Metric(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed otel reader; %w", err)
				}

				c.MetricReaders.Otel = otelReader
				readers = append(readers, otelReader)
			case types.MetricPrometheus:
				prometheusReader := exporter.Prometheus{}.Metric()
				c.MetricReaders.Prometheus = prometheusReader
				readers = append(readers, prometheusReader)
			}
		}

		c.MetricProvider(views, readers...).SetMetricProviderGlobal()
	}

	tracesEnabled := cfg.GetEnabledTraces()
	if len(tracesEnabled) > 0 {
		// set metrics
		for _, v := range tracesEnabled {
			switch v {
			case types.TraceOtel:
				if err := c.TraceProvider(ctx); err != nil {
					return nil, err
				}
			}
		}

		c.SetTraceProviderGlobal()
	}

	return c, nil
}

// Gets attributes key-value.
func (c *Collector) GetAttributes() []attribute.KeyValue {
	// add common attributes
	var attributes []attribute.KeyValue

	for k, v := range c.Attributes {
		attributes = append(attributes, attribute.String(strings.ToLower(k), fmt.Sprint(v)))
	}

	return attributes
}

// Shutdown to flush and shutdown providers and close grpc connection.
// Providers will not export metrics after shutdown.
func (c *Collector) Shutdown() (err error) {
	// set the default context timeout
	if c.ShutdownTimeOut == 0 {
		c.ShutdownTimeOut = defaultShutdownTimeOut
	}

	defer func() {
		if errClose := c.Conn.Close(); errClose != nil {
			err = fmt.Errorf("failed to close connection; %v; %w", errClose, err)
		}
	}()

	ctx, cancelCtx := context.WithTimeout(context.Background(), c.ShutdownTimeOut)
	defer cancelCtx()

	if errShutdown := c.MeterProvider.Shutdown(ctx); errShutdown != nil {
		err = fmt.Errorf("failed to shutdown meter provider; %w; %v", errShutdown, err)
	}

	if errShutdown := c.TracerProvider.Shutdown(ctx); errShutdown != nil {
		err = fmt.Errorf("failed to shutdown trace provider; %w; %v", errShutdown, err)
	}

	return
}
