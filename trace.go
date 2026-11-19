package tell

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
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
		return ErrSetConnection
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(c.Conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := tracesdk.NewBatchSpanProcessor(traceExporter)

	// Sampler gets from OTEL_TRACES_SAMPLER and OTEL_TRACES_SAMPLER_ARG, default is parentbased_always_on.
	c.TracerProviderSDK = tracesdk.NewTracerProvider(
		tracesdk.WithResource(c.Resource()),
		tracesdk.WithSpanProcessor(bsp),
	)

	c.TracerProvider = c.TracerProviderSDK

	return nil
}

// SetTraceProviderGlobal to globally set provider.
func (c *Collector) SetTraceProviderGlobal() *Collector {
	otel.SetTracerProvider(c.TracerProvider)

	if !c.isTraceNoop {
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			),
		)
	}

	return c
}
