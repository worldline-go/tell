package tell

import (
	"go.opentelemetry.io/otel/sdk/resource"
)

// Resource returns the shared resource of metric and trace providers.
//
// It has the SDK's default attributes (telemetry.sdk.*, service.name)
// with OTEL_SERVICE_NAME and OTEL_RESOURCE_ATTRIBUTES environment values.
func (c *Collector) Resource() *resource.Resource {
	return resource.Default()
}
