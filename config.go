package tell

import "time"

type Config struct {
	// Collector to show URL of grpc otel collector.
	// If emptry disable for metric and trace. It is add a noop metric/trace and your code works without change.
	Collector string         `cfg:"collector"`
	Metric    MetricSettings `cfg:"metric"`
	Trace     TraceSettings  `cfg:"trace"`
}

type MetricSettings struct {
	Provider MetricProviderSettings `cfg:"provider"`
	Disable  bool                   `cfg:"disable"`
}

type MetricProviderSettings struct {
	Interval time.Duration `cfg:"interval"`
}

type TraceSettings struct {
	Provider TraceProviderSettings `cfg:"provider"`
	Disable  bool                  `cfg:"disable"`
}

type TraceProviderSettings struct{}
