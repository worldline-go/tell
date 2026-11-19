package tell

import (
	"net/url"
	"os"
	"strconv"
	"strings"
)

const (
	envEndpoint          = "OTEL_EXPORTER_OTLP_ENDPOINT"
	envInsecure          = "OTEL_EXPORTER_OTLP_INSECURE"
	envCertificate       = "OTEL_EXPORTER_OTLP_CERTIFICATE"
	envClientCertificate = "OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE"
	envClientKey         = "OTEL_EXPORTER_OTLP_CLIENT_KEY"
	envCompression       = "OTEL_EXPORTER_OTLP_COMPRESSION"
	envProtocol          = "OTEL_EXPORTER_OTLP_PROTOCOL"
	envSDKDisabled       = "OTEL_SDK_DISABLED"
	envMetricsExporter   = "OTEL_METRICS_EXPORTER"
	envTracesExporter    = "OTEL_TRACES_EXPORTER"
)

// applyEnv fills the configuration with OTEL_* environment values.
// Values set in the configuration have priority over the environment.
func (cfg *Config) applyEnv() {
	if cfg.Collector == "" {
		cfg.Collector = os.Getenv(envEndpoint)
	}

	if v, ok := envBool(envSDKDisabled); ok && v {
		cfg.Metric.Disable = true
		cfg.Trace.Disable = true
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv(envMetricsExporter)), "none") {
		cfg.Metric.Disable = true
	}

	if strings.EqualFold(strings.TrimSpace(os.Getenv(envTracesExporter)), "none") {
		cfg.Trace.Disable = true
	}

	if v, ok := envBool(envInsecure); ok && !v {
		cfg.TLS.Enabled = true
	}

	if cfg.TLS.CAFile == "" {
		cfg.TLS.CAFile = os.Getenv(envCertificate)
	}

	if cfg.TLS.CertFile == "" && cfg.TLS.KeyFile == "" {
		cfg.TLS.CertFile = os.Getenv(envClientCertificate)
		cfg.TLS.KeyFile = os.Getenv(envClientKey)
	}

	if cfg.Compression == "" {
		cfg.Compression = strings.ToLower(strings.TrimSpace(os.Getenv(envCompression)))
	}

	if cfg.Protocol == "" {
		cfg.Protocol = os.Getenv(envProtocol)
	}

	cfg.Collector = cfg.parseEndpoint(cfg.Collector)
}

// parseEndpoint returns the host of the endpoint.
//   - "https://host:4317" enables TLS and returns "host:4317".
//   - "http://host:4317" returns "host:4317" without TLS.
//   - Path of the URL is kept in basePath for HTTP protocol, e.g. "https://host/otlp" sends to "/otlp/v1/traces".
//   - Other values are returned as they are, e.g. "host:4317" or "dns:///host:4317".
func (cfg *Config) parseEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		return endpoint
	}

	switch strings.ToLower(u.Scheme) {
	case "https":
		cfg.TLS.Enabled = true
	case "http":
	default:
		return endpoint
	}

	cfg.basePath = strings.TrimRight(u.Path, "/")

	return u.Host
}

func envBool(key string) (value, ok bool) {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return false, false
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, false
	}

	return b, true
}

// validEnvMillis reports whether the env value is a positive duration in milliseconds as the SDK expects.
func validEnvMillis(key string) bool {
	v, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(key)), 10, 64)

	return err == nil && v > 0
}
