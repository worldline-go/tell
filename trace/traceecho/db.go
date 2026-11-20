package traceecho

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DBMiddleware for tracing only database middlewares.
//   - Use this after echo tracing middleware.
func DBMiddleware(dbName, spanName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			ctx := c.Request().Context()
			ctx, span := otel.Tracer("").Start(ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(attribute.String("db.name", dbName)),
			)
			defer span.End()
			defer func() {
				// default error status code
				if _, status := echo.ResolveResponseStatus(c.Response(), err); status >= http.StatusInternalServerError {
					span.SetStatus(codes.Error, http.StatusText(status))
				}
			}()

			// set context to request
			c.SetRequest(c.Request().WithContext(ctx))

			// add span to change status code and message
			c.Set("traceecho-db-span", span)

			return next(c)
		}
	}
}
