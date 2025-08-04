package tracehttp

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Middleware(opts ...otelhttp.Option) func(next http.Handler) http.Handler {
	return otelhttp.NewMiddleware("", append([]otelhttp.Option{
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return strings.ToUpper(r.Method) + " " + r.URL.Path
		}),
	}, opts...)...)
}
