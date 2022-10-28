package tell

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/view"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/config"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/metric/exporter"
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
	MeterProvider    metric.MeterProvider
	MeterProviderSDK *metricsdk.MeterProvider
	MetricReaders    MetricReaders
	// traces
	TracerProvider    trace.TracerProvider
	TracerProviderSDK *tracesdk.TracerProvider
	// ShutdownTimeOut for closing providers, default 2 seconds.
	ShutdownTimeOut time.Duration
}

type MetricReaders struct {
	Otel       metricsdk.Reader
	Prometheus *prometheus.Exporter
}

// New generate collectors based on configuration.
func New(ctx context.Context, cfg Config, views ...view.View) (*Collector, error) {
	if cfg.Collector == "" {
		cfg.Collector = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	if cfg.Collector != "" {
		log.Info().Msgf("opentelemetry collector: %s", cfg.Collector)
	}

	c := new(Collector)
	c.Attributes = cfg.Attributes

	// check grpc need
	if cfg.Collector != "" && cfg.IsGRPC() {
		if err := c.ConnectGRPC(ctx, cfg.Collector); err != nil {
			return nil, err
		}

		log.Info().Msg("connected to grpc opentelemetry collector")
	}

	// metricsViewEnabled := cfg.GetEnabledViews()
	// if len(metricsViewEnabled) > 0 && !cfg.Disable {
	// 	// set enabled metric views here and append to views
	// }

	metricsEnabled := cfg.GetEnabledMetrics()
	if cfg.Collector != "" && len(metricsEnabled) > 0 {
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

				log.Info().Msg("started metric provider for otel")
			case types.MetricPrometheus:
				prometheusReader := exporter.Prometheus{}.Metric()
				c.MetricReaders.Prometheus = prometheusReader
				readers = append(readers, prometheusReader)

				log.Info().Msg("started metric provider for prometheus")
			}
		}

		c.MetricProvider(views, readers...).SetMetricProviderGlobal()
	} else {
		c.MeterProvider = metric.NewNoopMeterProvider()
		c.SetMetricProviderGlobal()

		log.Info().Msg("started metric provider noop")
	}

	tracesEnabled := cfg.GetEnabledTraces()
	if cfg.Collector != "" && len(tracesEnabled) > 0 {
		// set metrics
		for _, v := range tracesEnabled {
			switch v {
			case types.TraceOtel:
				if err := c.TraceProvider(ctx); err != nil {
					return nil, err
				}

				log.Info().Msg("started trace provider for otel")
			}
		}

		c.SetTraceProviderGlobal()
	} else {
		c.TracerProvider = trace.NewNoopTracerProvider()
		c.SetTraceProviderGlobal()

		log.Info().Msg("started trace provider noop")
	}

	return c, nil
}

// Gets attributes key-value.
func (c *Collector) GetAttributes() []attribute.KeyValue {
	// add common attributes
	var attributes []attribute.KeyValue //nolint:prealloc // return nil on empty

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
		if c.Conn != nil {
			if errClose := c.Conn.Close(); errClose != nil {
				err = fmt.Errorf("failed to close connection; %v; %w", errClose, err)
			}
		}
	}()

	ctx, cancelCtx := context.WithTimeout(context.Background(), c.ShutdownTimeOut)
	defer cancelCtx()

	if c.MeterProviderSDK != nil {
		if errShutdown := c.MeterProviderSDK.Shutdown(ctx); errShutdown != nil {
			err = fmt.Errorf("failed to shutdown meter provider; %w; %v", errShutdown, err)
		}
	}

	if c.TracerProviderSDK != nil {
		if errShutdown := c.TracerProviderSDK.Shutdown(ctx); errShutdown != nil {
			err = fmt.Errorf("failed to shutdown trace provider; %w; %v", errShutdown, err)
		}
	}

	return
}
