package tell

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	ProtocolGRPC         = "grpc"
	ProtocolHTTPProtobuf = "http/protobuf"
)

// exporterConfig holds the settings of the HTTP exporters.
// Zero value uses gRPC exporters with the collector's connection.
type exporterConfig struct {
	http        bool
	endpoint    string
	basePath    string
	tls         *tls.Config
	compression bool
}

// normalizeProtocol returns the supported protocol name.
func normalizeProtocol(protocol string) (string, error) {
	switch p := strings.ToLower(strings.TrimSpace(protocol)); p {
	case "", ProtocolGRPC:
		return ProtocolGRPC, nil
	case ProtocolHTTPProtobuf:
		return p, nil
	default:
		return "", fmt.Errorf("unsupported protocol [%s], use %s or %s", protocol, ProtocolGRPC, ProtocolHTTPProtobuf)
	}
}

// httpExporterConfig returns the HTTP exporter settings of the configuration.
func (cfg Config) httpExporterConfig() (exporterConfig, error) {
	compression, err := useGzip(cfg.Compression)
	if err != nil {
		return exporterConfig{}, err
	}

	e := exporterConfig{
		http:        true,
		endpoint:    cfg.Collector,
		basePath:    cfg.basePath,
		compression: compression,
	}

	if cfg.TLS.Enabled {
		tlsConfig, err := cfg.TLS.Generate()
		if err != nil {
			return exporterConfig{}, fmt.Errorf("failed to generate tls config; %w", err)
		}

		e.tls = tlsConfig
	}

	return e, nil
}

func useGzip(compression string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(compression)) {
	case "", "none":
		return false, nil
	case "gzip":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported compression [%s], use gzip or none", compression)
	}
}

func (c *Collector) metricExporter(ctx context.Context) (metricsdk.Exporter, error) {
	if !c.exporter.http {
		if c.Conn == nil {
			return nil, ErrSetConnection
		}

		return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(c.Conn))
	}

	e := c.exporter
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(e.endpoint),
		otlpmetrichttp.WithURLPath(e.basePath + "/v1/metrics"),
	}

	if e.tls != nil {
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(e.tls))
	} else {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	// always set, otherwise the SDK reads OTEL_EXPORTER_OTLP_COMPRESSION and overrides the config
	compression := otlpmetrichttp.NoCompression
	if e.compression {
		compression = otlpmetrichttp.GzipCompression
	}

	opts = append(opts, otlpmetrichttp.WithCompression(compression))

	return otlpmetrichttp.New(ctx, opts...)
}

func (c *Collector) traceExporter(ctx context.Context) (tracesdk.SpanExporter, error) {
	if !c.exporter.http {
		if c.Conn == nil {
			return nil, ErrSetConnection
		}

		return otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(c.Conn))
	}

	e := c.exporter
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(e.endpoint),
		otlptracehttp.WithURLPath(e.basePath + "/v1/traces"),
	}

	if e.tls != nil {
		opts = append(opts, otlptracehttp.WithTLSClientConfig(e.tls))
	} else {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	compression := otlptracehttp.NoCompression
	if e.compression {
		compression = otlptracehttp.GzipCompression
	}

	opts = append(opts, otlptracehttp.WithCompression(compression))

	return otlptracehttp.New(ctx, opts...)
}
