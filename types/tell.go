package types

import "strings"

type Tell int64

const (
	Unknown Tell = iota
	MetricOtel
	MetricPrometheus
	ViewRequestDuration
	TraceOtel
)

func (t Tell) String() string {
	switch t {
	case MetricOtel:
		return "metric_otel"
	case MetricPrometheus:
		return "metric_prometheus"
	case ViewRequestDuration:
		return "view_request_duration"
	case TraceOtel:
		return "trace_otel"
	}

	return "unknown"
}

type Tells []Tell

func (t Tells) Strings(trimPrefix string) []string {
	tells := make([]string, len(t))

	for i, v := range t {
		tells[i] = strings.TrimLeft(v.String(), trimPrefix)
	}

	return tells
}

func (t Tells) Map() map[Tell]struct{} {
	tMap := make(map[Tell]struct{}, len(t))

	for _, v := range t {
		tMap[v] = struct{}{}
	}

	return tMap
}

func (t Tells) IsExistOne(tChecks ...Tells) bool {
	tMap := t.Map()

	for _, tCheck := range tChecks {
		for _, v := range tCheck {
			if _, ok := tMap[v]; ok {
				return true
			}
		}
	}

	return false
}

func StringToTell(v string) Tell {
	switch v {
	case "metric_otel":
		return MetricOtel
	case "metric_prometheus":
		return MetricPrometheus
	case "view_request_duration":
		return ViewRequestDuration
	case "trace_otel":
		return TraceOtel
	}

	return Unknown
}

func SliceToTells(v []string, prefix string) Tells {
	tells := make(Tells, len(v))

	for i, v := range v {
		tells[i] = StringToTell(prefix + v)
	}

	return tells
}
