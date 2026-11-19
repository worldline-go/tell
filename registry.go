package tell

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding/gzip"
)

var (
	// ErrSetConnection is returned when provider initialization requires a gRPC connection.
	ErrSetConnection = errors.New("grpc connection not set")
	// ErrSetConnetion is kept for backward compatibility.
	//
	// Deprecated: use ErrSetConnection.
	ErrSetConnetion = ErrSetConnection
)

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

	exporter exporterConfig

	isUp       atomic.Int64
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

		return
	}

	regUp, err := meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		o.ObserveInt64(up, c.isUp.Load())

		return nil
	}, up)

	if err != nil {
		c.logger.Error("failed to register up gauge metric", "error", err.Error())
	}

	c.AddRegister(regUp)
}

// AddRegister adding metric.Registration for unregister in shutdown.
func (c *Collector) AddRegister(r metric.Registration) {
	if r == nil {
		return
	}

	c.registered = append(c.registered, r)
}

// New generate collectors based on configuration.
//
// opts are only used with grpc protocol.
//
// Empty configuration values are filled with OTEL_* environment values, check the Config fields.
func New(ctx context.Context, cfg Config, opts ...grpc.DialOption) (*Collector, error) {
	cfg.applyEnv()

	logger := cfg.Logger
	if logger == nil {
		logger = adapterKV{Log: log.Logger}
	}

	c := &Collector{
		logger: logger,
	}

	useMetric := cfg.Collector != "" && !cfg.Metric.Disable
	useTrace := cfg.Collector != "" && !cfg.Trace.Disable

	if useMetric || useTrace {
		protocol, err := normalizeProtocol(cfg.Protocol)
		if err != nil {
			return nil, err
		}

		c.logger.Info(fmt.Sprintf("opentelemetry collector endpoint: [%s] protocol: [%s]", cfg.Collector, protocol))

		switch protocol {
		case ProtocolHTTPProtobuf:
			if c.exporter, err = cfg.httpExporterConfig(); err != nil {
				return nil, err
			}

			if len(opts) > 0 || cfg.ServerName != "" {
				c.logger.Warn("grpc dial options and server_name are ignored with http protocol")
			}
		default:
			dialOpts, err := cfg.dialOptions()
			if err != nil {
				return nil, err
			}

			// user options have priority
			if err := c.ConnectGRPC(ctx, cfg.Collector, append(dialOpts, opts...)...); err != nil {
				return nil, err
			}

			c.logger.Info("connected to grpc opentelemetry collector")
		}
	}

	// metric
	if useMetric {
		if err := c.MetricProvider(ctx, cfg.Metric.Provider); err != nil {
			return nil, errors.Join(fmt.Errorf("failed initialize metric provider; %w", err), c.Shutdown())
		}

		c.logger.Info("started metric provider for [otel]")

		// add enabled metrics
		if cfg.Metric.Default.GoRuntime {
			if err := runtime.Start(runtime.WithMeterProvider(c.MeterProvider)); err != nil {
				return nil, errors.Join(fmt.Errorf("failed to start runtime metrics; %w", err), c.Shutdown())
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
	if useTrace {
		if err := c.TraceProvider(ctx, cfg.Trace.Provider); err != nil {
			return nil, errors.Join(fmt.Errorf("failed initialize trace provider; %w", err), c.Shutdown())
		}

		c.logger.Info("started trace provider for [otel]")
	} else {
		c.TracerProvider = traceNoop.NewTracerProvider()
		c.isTraceNoop = true
		c.logger.Info("started trace provider for [noop]")
	}

	c.SetTraceProviderGlobal()

	// everything is works fine, send up information
	c.isUp.Store(1)
	c.setUpMetric()

	return c, nil
}

// dialOptions returns grpc options for TLS, server name and compression.
func (cfg Config) dialOptions() ([]grpc.DialOption, error) {
	var opts []grpc.DialOption

	if cfg.TLS.Enabled {
		tlsConfig, err := cfg.TLS.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate tls config; %w", err)
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if cfg.ServerName != "" {
		opts = append(opts, grpc.WithAuthority(cfg.ServerName))
	}

	compression, err := useGzip(cfg.Compression)
	if err != nil {
		return nil, err
	}

	if compression {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	}

	return opts, nil
}

// Shutdown to flush and shutdown providers and close grpc connection.
// Providers will not export metrics after shutdown.
func (c *Collector) Shutdown() (err error) {
	c.isUp.Store(0)

	// set the default context timeout
	if c.ShutdownTimeOut == 0 {
		c.ShutdownTimeOut = defaultShutdownTimeOut
	}

	defer func() {
		if errClose := c.CloseGRPC(); errClose != nil {
			err = errors.Join(err, errClose)
		}

		c.Conn = nil
	}()

	ctxMetric, cancelCtxMetric := context.WithTimeout(context.Background(), c.ShutdownTimeOut)
	defer cancelCtxMetric()

	if c.MeterProviderSDK != nil {
		if errShutdown := c.MeterProviderSDK.Shutdown(ctxMetric); errShutdown != nil {
			err = errors.Join(err, fmt.Errorf("failed to shutdown meter provider; %w", errShutdown))
		}

		c.MeterProviderSDK = nil
	}

	ctxTrace, cancelCtxTrace := context.WithTimeout(context.Background(), c.ShutdownTimeOut)
	defer cancelCtxTrace()

	if c.TracerProviderSDK != nil {
		if errShutdown := c.TracerProviderSDK.Shutdown(ctxTrace); errShutdown != nil {
			err = errors.Join(err, fmt.Errorf("failed to shutdown trace provider; %w", errShutdown))
		}

		c.TracerProviderSDK = nil
	}

	// remove registration
	for _, r := range c.registered {
		if errUnregister := r.Unregister(); errUnregister != nil {
			err = errors.Join(err, fmt.Errorf("failed to unregister metric callback; %w", errUnregister))
		}
	}

	c.registered = nil

	return err
}
