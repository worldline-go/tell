package tracehttp

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// DBMiddleware for tracing only database middlewares.
//   - Use this after tracing middleware.
func DBMiddleware(dbName, spanName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx, span := otel.Tracer("").Start(ctx,
				spanName,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(attribute.String("db.name", dbName)),
			)

			w2 := &responseWriter{ResponseWriter: w}

			defer span.End()
			defer func() {
				// default error status code
				if w2.status >= http.StatusInternalServerError {
					span.SetStatus(codes.Error, http.StatusText(w2.status))
				}
			}()

			next.ServeHTTP(w2, r.WithContext(ctx))
		})
	}
}
