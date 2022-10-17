package exporter

import "go.opentelemetry.io/otel/exporters/prometheus"

// MetricPrometheus return a reader wuth prometheus.
func MetricPrometheus() prometheus.Exporter {
	exporter := prometheus.New()

	return exporter
}
