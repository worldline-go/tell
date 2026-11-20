package metricecho

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
)

// HTTPMetrics is an echo middleware to add metrics to rec for each HTTP request.
// If recorder config is nil, the middleware will use a recorder with default configuration.
func HTTPMetrics(opts ...Option) echo.MiddlewareFunc {
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

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			values := HTTPLabels{
				Method: c.Request().Method,
				Path:   c.Path(),
			}
			ctx := context.WithoutCancel(c.Request().Context())

			rec.AddInFlightRequest(ctx, values)

			start := time.Now()

			defer func() {
				elapsed := time.Since(start)

				_, values.Code = echo.ResolveResponseStatus(c.Response(), err)

				rec.AddRequestToTotal(ctx, values)
				rec.AddRequestDuration(ctx, elapsed, values)
				rec.RemInFlightRequest(ctx, values)
			}()

			return next(c)
		}
	}
}

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
