package tell

import (
	"testing"
)

func TestApplyEnv(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		env  map[string]string
		want func(t *testing.T, cfg Config)
	}{
		{
			name: "endpoint from env with https",
			env:  map[string]string{envEndpoint: "https://otel-collector:4317"},
			want: func(t *testing.T, cfg Config) {
				if cfg.Collector != "otel-collector:4317" {
					t.Errorf("collector = %q", cfg.Collector)
				}
				if !cfg.TLS.Enabled {
					t.Error("tls should be enabled")
				}
			},
		},
		{
			name: "endpoint with http",
			cfg:  Config{Collector: "http://otel-collector:4317"},
			want: func(t *testing.T, cfg Config) {
				if cfg.Collector != "otel-collector:4317" {
					t.Errorf("collector = %q", cfg.Collector)
				}
				if cfg.TLS.Enabled {
					t.Error("tls should be disabled")
				}
			},
		},
		{
			name: "endpoint without scheme",
			cfg:  Config{Collector: "otel-collector:4317"},
			want: func(t *testing.T, cfg Config) {
				if cfg.Collector != "otel-collector:4317" {
					t.Errorf("collector = %q", cfg.Collector)
				}
			},
		},
		{
			name: "config has priority",
			cfg:  Config{Collector: "config:4317", Compression: "none"},
			env:  map[string]string{envEndpoint: "env:4317", envCompression: "gzip"},
			want: func(t *testing.T, cfg Config) {
				if cfg.Collector != "config:4317" {
					t.Errorf("collector = %q", cfg.Collector)
				}
				if cfg.Compression != "none" {
					t.Errorf("compression = %q", cfg.Compression)
				}
			},
		},
		{
			name: "protocol from env",
			env:  map[string]string{envProtocol: "http/protobuf"},
			want: func(t *testing.T, cfg Config) {
				if cfg.Protocol != "http/protobuf" {
					t.Errorf("protocol = %q", cfg.Protocol)
				}
			},
		},
		{
			name: "endpoint with path",
			cfg:  Config{Collector: "https://otel.example.com/otlp/"},
			want: func(t *testing.T, cfg Config) {
				if cfg.Collector != "otel.example.com" || cfg.basePath != "/otlp" {
					t.Errorf("collector = %q, basePath = %q", cfg.Collector, cfg.basePath)
				}
			},
		},
		{
			name: "sdk disabled",
			env:  map[string]string{envSDKDisabled: "true"},
			want: func(t *testing.T, cfg Config) {
				if !cfg.Metric.Disable || !cfg.Trace.Disable {
					t.Error("metric and trace should be disabled")
				}
			},
		},
		{
			name: "exporters none",
			env:  map[string]string{envMetricsExporter: "none"},
			want: func(t *testing.T, cfg Config) {
				if !cfg.Metric.Disable {
					t.Error("metric should be disabled")
				}
				if cfg.Trace.Disable {
					t.Error("trace should be enabled")
				}
			},
		},
		{
			name: "tls from env",
			env: map[string]string{
				envInsecure:          "false",
				envCertificate:       "/ca.pem",
				envClientCertificate: "/cert.pem",
				envClientKey:         "/key.pem",
			},
			want: func(t *testing.T, cfg Config) {
				if !cfg.TLS.Enabled {
					t.Error("tls should be enabled")
				}
				if cfg.TLS.CAFile != "/ca.pem" || cfg.TLS.CertFile != "/cert.pem" || cfg.TLS.KeyFile != "/key.pem" {
					t.Errorf("tls = %+v", cfg.TLS)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range []string{
				envEndpoint, envInsecure, envCertificate, envClientCertificate, envClientKey,
				envCompression, envSDKDisabled, envMetricsExporter, envTracesExporter, envProtocol,
			} {
				t.Setenv(k, "")
			}

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			cfg := tt.cfg
			cfg.applyEnv()
			tt.want(t, cfg)
		})
	}
}

func TestNewNoop(t *testing.T) {
	t.Setenv(envEndpoint, "")

	c, err := New(t.Context(), Config{})
	if err != nil {
		t.Fatal(err)
	}

	if !c.IsMetricNoop() || !c.IsTraceNoop() {
		t.Error("providers should be noop")
	}

	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}
}

func TestValidEnvMillis(t *testing.T) {
	for v, want := range map[string]bool{"": false, "5000": true, "5s": false, "0": false, "-1": false} {
		t.Setenv(envMetricExportInterval, v)

		if got := validEnvMillis(envMetricExportInterval); got != want {
			t.Errorf("%q = %v, want %v", v, got, want)
		}
	}
}

func TestShutdownTwice(t *testing.T) {
	t.Setenv(envEndpoint, "localhost:4317")

	// metric export fails without a collector, trace has nothing to export
	c, err := New(t.Context(), Config{Metric: MetricSettings{Disable: true}})
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}

	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}
}
