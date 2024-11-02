package traceecho

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Span get span from echo context.
//   - Use this after echo tracing middleware.
//   - If span not found, return nil.
func Span(c echo.Context) trace.Span {
	span, _ := c.Get("span").(trace.Span)

	return span
}

// DBMiddleware for tracing only database middlewares.
//   - Use this after echo tracing middleware.
func DBMiddleware(dbName, spanName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ctx := c.Request().Context()
			ctx, span := otel.Tracer("").Start(ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(attribute.String("db.name", dbName)),
			)
			defer span.End()
			defer func() {
				// default error status code
				if c.Response().Status >= 500 {
					span.SetStatus(codes.Error, http.StatusText(c.Response().Status))
				}
			}()

			defer func() {
				if err != nil {
					c.Error(err)
					// don't return the error so that it's not handled again
					err = nil
				}
			}()

			// set context to request
			c.SetRequest(c.Request().WithContext(ctx))

			// add span to change status code and message
			c.Set("span", span)

			return next(c)
		}
	}
}
