package config

import (
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/metric/exporter"
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/types"
)

var (
	// Problem on multi readers
	// VMetricAll = types.Tells{types.MetricOtel, types.MetricPrometheus}
	VMetricAll = types.Tells{types.MetricOtel}
	VViewAll   = types.Tells{}
	VGrpcNeed  = types.Tells{types.MetricOtel, types.TraceOtel}

	VTraceAll = types.Tells{types.TraceOtel}
)

type Config struct {
	// Attributes have common attributes.
	Attributes map[string]interface{}

	Traces Selectors `cfg:"traces"`

	Metrics         Selectors       `cfg:"metrics"`
	MetricsViews    Selectors       `cfg:"metrics_views"`
	MetricsSettings MetricsSettings `cfg:"metrics_settings"`
	// Collector to show URL of grpc otel collector.
	Collector string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"otel-collector:4317"`
	// Disable for metric and trace. It is add a noop metric/trace and your code works without change.
	Disable bool
}

func (c *Config) GetEnabledViews() types.Tells {
	return types.SliceToTells(c.MetricsViews.GetEnabled(VViewAll.Strings("view")), "view")
}

func (c *Config) GetEnabledMetrics() types.Tells {
	return types.SliceToTells(c.Metrics.GetEnabled(VMetricAll.Strings("metric")), "metric")
}

func (c *Config) GetEnabledTraces() types.Tells {
	return types.SliceToTells(c.Metrics.GetEnabled(VTraceAll.Strings("trace")), "trace")
}

func (c *Config) IsGRPC() bool {
	return VGrpcNeed.IsExistOne(c.GetEnabledMetrics()) || VGrpcNeed.IsExistOne(c.GetEnabledTraces())
}

type Selectors struct {
	Enable     []string
	Disable    []string
	EnableAll  bool `cfg:"enable_all" default:"true"`
	DisableAll bool `cfg:"disable_all"`
}

func (s *Selectors) GetEnabled(all []string) []string {
	if s.DisableAll {
		return nil
	}

	selected := s.Enable

	if s.EnableAll {
		selected = all
	}

	if s.Disable == nil {
		return selected
	}

	// slice to map
	selectedMap := map[string]struct{}{}
	for _, v := range selected {
		selectedMap[v] = struct{}{}
	}

	for _, v := range s.Disable {
		delete(selectedMap, v)
	}

	// map to slice
	selected = make([]string, 0, len(selectedMap))

	for key := range selectedMap {
		selected = append(selected, key)
	}

	return selected
}

type MetricsSettings struct {
	Otel exporter.OtelSetting
}
