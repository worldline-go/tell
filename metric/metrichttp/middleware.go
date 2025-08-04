package metrichttp

import (
	"context"
	"net/http"
	"time"
)

// Middleware is an echo middleware to add metrics to rec for each HTTP request.
// If recorder config is nil, the middleware will use a recorder with default configuration.
func Middleware(opts ...Option) func(next http.Handler) http.Handler {
	option := option{
		cfg: HTTPCfg,
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}

		opt(&option)
	}

	rec := NewHTTPRecorder(option.cfg, nil)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			values := HTTPLabels{
				Method: r.Method,
				Path:   r.URL.Path,
			}

			ctx := context.WithoutCancel(r.Context())

			rec.AddInFlightRequest(ctx, values)

			start := time.Now()

			w2 := &responseWriter{ResponseWriter: w}

			defer func() {
				elapsed := time.Since(start)

				values.Code = w2.status

				rec.AddRequestToTotal(ctx, values)
				rec.AddRequestDuration(ctx, elapsed, values)
				rec.RemInFlightRequest(ctx, values)
			}()

			next.ServeHTTP(w2, r)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// ///////////////////////////////////////////////////

type option struct {
	cfg HTTPRecorderConfig
}

type Option func(*option)

func WithHTTPRecorderConfig(cfg HTTPRecorderConfig) Option {
	return func(o *option) {
		o.cfg = cfg
	}
}

func WithTotalMetric(v bool) Option {
	return func(o *option) {
		o.cfg.EnableTotalMetric = v
	}
}

func WithDurMetric(v bool) Option {
	return func(o *option) {
		o.cfg.EnableDurMetric = v
	}
}

func WithInFlightMetric(v bool) Option {
	return func(o *option) {
		o.cfg.EnableInFlightMetric = v
	}
}

func WithNamespace(v string) Option {
	return func(o *option) {
		o.cfg.Namespace = v
	}
}
