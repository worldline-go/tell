package tell

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	metricNoop "go.opentelemetry.io/otel/metric/noop"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	traceNoop "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var ErrSetConnetion = errors.New("grpc connection not set")

const defaultShutdownTimeOut = 2 * time.Second

// Collector hold metric and trace informations.
type Collector struct {
	Conn *grpc.ClientConn
	// metrics
	MeterProvider    metric.MeterProvider
	MeterProviderSDK *metricsdk.MeterProvider
	MetricReader     metricsdk.Reader
	isMetricNoop     bool
	// traces
	TracerProvider    trace.TracerProvider
	TracerProviderSDK *tracesdk.TracerProvider
	isTraceNoop       bool
	// ShutdownTimeOut for closing providers, default 2 seconds.
	ShutdownTimeOut time.Duration

	isUp       int64
	registered []metric.Registration
	logger     Logger
}

func (c *Collector) IsMetricNoop() bool {
	return c.isMetricNoop
}

func (c *Collector) IsTraceNoop() bool {
	return c.isTraceNoop
}

func (c *Collector) setUpMetric() {
	meter := c.MeterProvider.Meter("")

	up, err := meter.Int64ObservableGauge("up", metric.WithDescription("application up status"))
	if err != nil {
		c.logger.Error("failed to set up gauge metric", "error", err.Error())
	}

	regUp, err := meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveInt64(up, c.isUp)

		return nil
	}, up)

	if err != nil {
		c.logger.Error("failed to register up gauge metric", "error", err.Error())
	}

	c.AddRegister(regUp)
}

// AddRegister adding metric.Registration for unregister in shutdown.
func (c *Collector) AddRegister(r metric.Registration) {
	c.registered = append(c.registered, r)
}

// New generate collectors based on configuration.
func New(ctx context.Context, cfg Config, opts ...grpc.DialOption) (*Collector, error) {
	if cfg.Collector == "" {
		cfg.Collector = os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = adapterKV{Log: log.Logger}
	}

	c := &Collector{
		logger: logger,
	}

	if cfg.Collector != "" {
		c.logger.Info(fmt.Sprintf("opentelemetry collector endpoint: [%s]", cfg.Collector))
	}

	// check grpc need
	if cfg.Collector != "" {
		if cfg.TLS.Enabled {
			tlsConfig, err := cfg.TLS.Generate()
			if err != nil {
				return nil, fmt.Errorf("failed to generate tls config; %w", err)
			}

			opts = append([]grpc.DialOption{
				grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
			}, opts...)
		}

		if cfg.ServerName != "" {
			opts = append([]grpc.DialOption{grpc.WithAuthority(cfg.ServerName)}, opts...)
		}

		if err := c.ConnectGRPC(ctx, cfg.Collector, opts...); err != nil {
			return nil, err
		}

		c.logger.Info("connected to grpc opentelemetry collector")
	}

	// metric
	if cfg.Collector != "" && !cfg.Metric.Disable {
		if err := c.MetricProvider(ctx, cfg.Metric.Provider); err != nil {
			return nil, fmt.Errorf("failed initialize metric provider; %w", err)
		}

		c.logger.Info("started metric provider for [otel]")

		// add enabled metrics
		if cfg.Metric.Default.GoRuntime {
			if err := runtime.Start(); err != nil {
				return nil, fmt.Errorf("failed to start runtime metrics; %w", err)
			}

			c.logger.Info("started runtime metrics")
		}
	} else {
		c.MeterProvider = metricNoop.NewMeterProvider()
		c.isMetricNoop = true
		c.logger.Info("started metric provider for [noop]")
	}

	c.SetMetricProviderGlobal()

	// trace
	if cfg.Collector != "" && !cfg.Trace.Disable {
		if err := c.TraceProvider(ctx, cfg.Trace.Provider); err != nil {
			return nil, fmt.Errorf("failed initialize metric provider; %w", err)
		}

		c.logger.Info("started trace provider for [otel]")
	} else {
		c.TracerProvider = traceNoop.NewTracerProvider()
		c.isTraceNoop = true
		c.logger.Info("started trace provider for [noop]")
	}

	c.SetTraceProviderGlobal()

	// everything is works fine, send up information
	c.isUp = 1
	c.setUpMetric()

	return c, nil
}

// Shutdown to flush and shutdown providers and close grpc connection.
// Providers will not export metrics after shutdown.
func (c *Collector) Shutdown() (err error) {
	c.isUp = 0

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

	ctxMetric, cancelCtxMetric := context.WithTimeout(context.Background(), c.ShutdownTimeOut)
	defer cancelCtxMetric()

	if c.MeterProviderSDK != nil {
		if errShutdown := c.MeterProviderSDK.Shutdown(ctxMetric); errShutdown != nil {
			err = fmt.Errorf("failed to shutdown meter provider; %w; %v", errShutdown, err)
		}
	}

	ctxTrace, cancelCtxTrace := context.WithTimeout(context.Background(), c.ShutdownTimeOut)
	defer cancelCtxTrace()

	if c.TracerProviderSDK != nil {
		if errShutdown := c.TracerProviderSDK.Shutdown(ctxTrace); errShutdown != nil {
			err = fmt.Errorf("failed to shutdown trace provider; %w; %v", errShutdown, err)
		}
	}

	// remove registiration
	for _, r := range c.registered {
		_ = r.Unregister()
	}

	return nil
}
