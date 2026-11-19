package tell

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
)

func TestNormalizeProtocol(t *testing.T) {
	for in, want := range map[string]string{"": ProtocolGRPC, "grpc": ProtocolGRPC, " HTTP/Protobuf ": ProtocolHTTPProtobuf} {
		got, err := normalizeProtocol(in)
		if err != nil || got != want {
			t.Errorf("%q = %q, %v; want %q", in, got, err, want)
		}
	}

	if _, err := normalizeProtocol("http/json"); err == nil {
		t.Error("http/json should be unsupported")
	}
}

func TestHTTPExport(t *testing.T) {
	t.Run("gzip", func(t *testing.T) { testHTTPExport(t, "gzip", "", true) })
	t.Run("config none has priority", func(t *testing.T) { testHTTPExport(t, "none", "gzip", false) })
}

func testHTTPExport(t *testing.T, compression, envCompressionValue string, wantGzip bool) {
	var (
		mu    sync.Mutex
		paths = map[string]string{}
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		paths[r.URL.Path] = r.Header.Get("Content-Encoding")
		mu.Unlock()

		w.Header().Set("Content-Type", "application/x-protobuf")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	for _, k := range []string{envEndpoint, envProtocol, envCompression, envSDKDisabled, envMetricsExporter, envTracesExporter, envInsecure} {
		t.Setenv(k, "")
	}

	t.Setenv(envCompression, envCompressionValue)

	c, err := New(t.Context(), Config{
		Collector:   srv.URL + "/otlp",
		Protocol:    ProtocolHTTPProtobuf,
		Compression: compression,
	})
	if err != nil {
		t.Fatal(err)
	}

	if c.Conn != nil {
		t.Error("grpc connection should not be opened for http protocol")
	}

	_, span := otel.Tracer("test").Start(context.Background(), "span")
	span.End()

	if err := c.Shutdown(); err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	defer mu.Unlock()

	for _, p := range []string{"/otlp/v1/metrics", "/otlp/v1/traces"} {
		enc, ok := paths[p]
		if !ok {
			t.Errorf("no request to %s, got %v", p, paths)

			continue
		}

		if strings.Contains(enc, "gzip") != wantGzip {
			t.Errorf("%s content-encoding = %q", p, enc)
		}
	}
}
