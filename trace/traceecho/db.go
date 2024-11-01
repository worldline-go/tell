package traceecho

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DBMiddleware for tracing only database middlewares.
//   - Use this after echo tracing middleware.
func DBMiddleware(dbName, spanName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			ctx, span := otel.Tracer("").Start(ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(attribute.String("db.name", dbName)),
			)
			defer span.End()
			defer func() {
				if c.Response().Status >= 500 {
					span.SetStatus(codes.Error, http.StatusText(c.Response().Status))
				}
			}()

			// set context to request
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}
