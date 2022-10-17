package metricecho

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
)

// HTTPMetrics is an echo middleware to add metrics to rec for each HTTP request. If rec is nil, the middleware wil
// use a recorder with default configuration.
//
// Add a namespace of application name, uses as prefix for metrics.
func HTTPMetrics(namespace string, rec *HTTPRecorder) echo.MiddlewareFunc {
	if rec == nil {
		rec = NewHTTPRecorder(HTTPCfg, nil)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			values := HTTPLabels{
				Method: c.Request().Method,
				Path:   c.Path(),
			}

			rec.AddInFlightRequest(context.Background(), values)

			start := time.Now()

			defer func() {
				elapsed := time.Since(start)

				if err != nil {
					c.Error(err)
					// don't return the error so that it's not handled again
					err = nil
				}

				values.Code = c.Response().Status

				rec.AddRequestToTotal(context.Background(), values)
				rec.AddRequestDuration(context.Background(), elapsed, values)
				rec.RemInFlightRequest(context.Background(), values)
			}()

			return next(c)
		}
	}
}
