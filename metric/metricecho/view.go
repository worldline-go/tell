package metricecho

import (
	"go.opentelemetry.io/otel/sdk/metric"
)

// DurationBuckets are the histogram boundaries used for request_duration_seconds.
var DurationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

// GetViews returns a view with the request duration buckets.
//
// Deprecated: buckets are set on the instrument directly, views are not needed anymore.
func GetViews() []metric.View {
	customBucketView := metric.NewView(
		metric.Instrument{
			Name: "*request_duration_seconds",
		},
		metric.Stream{
			Aggregation: metric.AggregationExplicitBucketHistogram{
				Boundaries: DurationBuckets,
			},
		},
	)

	return []metric.View{customBucketView}
}
