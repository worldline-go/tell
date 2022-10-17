package metricecho

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
)

var DefDurationBuckets = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}

func GetViews(durationBuckets []float64) ([]view.View, error) {
	if durationBuckets == nil {
		durationBuckets = DefDurationBuckets
	}

	customBucketsView, err := view.New(
		view.MatchInstrumentName("*request_duration_seconds"),
		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: durationBuckets,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("meter custom view cannot set; %w", err)
	}

	return []view.View{customBucketsView}, nil
}