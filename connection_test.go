package tell_test

import (
	"context"
	"testing"
	"time"

	"github.com/worldline-go/tell"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := tell.Collector{}

	header := grpc.Header(&metadata.MD{
		"Content-Type": []string{"application/grpc"},
	})

	if err := c.ConnectGRPC(ctx, "localhost:18080", grpc.WithBlock(), grpc.WithDefaultCallOptions(header)); err != nil {
		t.Fatal("failed to connect", err)
	}
}

func TestData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// open telemetry
	collector, err := tell.New(ctx,
		tell.Config{
			Collector: "localhost:443",
			TLS: tell.TLSConfig{
				Enabled:            true,
				InsecureSkipVerify: true,
			},
			ServerName: "otel-collector",
		},
		grpc.WithBlock(),
		// grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		// 	InsecureSkipVerify: true,
		// })),
		// grpc.WithAuthority("otel-collector"),
	)
	if err != nil {
		t.Fatal("failed to init telemetry", err)
	}

	defer collector.Shutdown()

	mp := otel.GetMeterProvider()

	meter := mp.Meter("")

	//nolint:lll // description
	testMeter, err := meter.Int64Counter("tell_connection_testing", metric.WithDescription("number of successfully connection count"))
	if err != nil {
		t.Fatal("failed to initialize tell_connection_testing", err)
	}

	testMeter.Add(ctx, 1)
}
