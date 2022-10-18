package exporter

import (
	"log"

	"go.opentelemetry.io/otel/exporters/prometheus"
)

type Prometheus struct{}

// MetricPrometheus return a reader as prometheus exporter.
func (m Prometheus) Metric() *prometheus.Exporter {
	exporter, err := prometheus.New()
	if err != nil {
		log.Fatal(err)
	}

	return exporter
}
