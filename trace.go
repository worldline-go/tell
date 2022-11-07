package tell

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// Get provider and don't forget to shutdown after usage, it will help to flush last messages.
// Use OTPL provider, it is more general and most of tools supporting and it is going to be standard.
//
// Also you can set globally this providor
// otel.SetTracerProvider(tracerProvider)
// set global propagator to tracecontext (the default is no-op).
// otel.SetTextMapPropagator(propagation.TraceContext{})

func (c *Collector) TraceProvider(ctx context.Context, _ TraceProviderSettings) error {
	if c.Conn == nil {
		return ErrSetConnetion
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(c.Conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := tracesdk.NewBatchSpanProcessor(traceExporter)

	// to much resources not need all of them this is just for example
	res, err := resource.New(ctx,
		// resource.WithContainer(),
		// resource.WithContainerID(),
		resource.WithFromEnv(),
		// resource.WithHost(),
		// resource.WithOS(),
		// resource.WithOSDescription(),
		// resource.WithOSType(),
		// resource.WithProcess(),
		// resource.WithProcessCommandArgs(),
		// resource.WithProcessExecutableName(),
		// resource.WithProcessExecutablePath(),
		// resource.WithProcessOwner(),
		// resource.WithProcessPID(),
		// resource.WithProcessRuntimeDescription(),
		// resource.WithProcessRuntimeName(),
		// resource.WithProcessRuntimeVersion(),
		// resource.WithTelemetrySDK(),
		resource.WithSchemaURL(semconv.SchemaURL),
		// resource.WithAttributes(
		// 	c.GetAttributes()...,
		// ),
	)
	if err != nil {
		return fmt.Errorf("failed resource new; %w", err)
	}

	c.TracerProviderSDK = tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithResource(res),
		tracesdk.WithSpanProcessor(bsp),
	)

	c.TracerProvider = c.TracerProviderSDK

	return nil
}

// SetTraceProviderGlobal to globally set provider.
func (c *Collector) SetTraceProviderGlobal() *Collector {
	otel.SetTracerProvider(c.TracerProvider)

	return c
}
