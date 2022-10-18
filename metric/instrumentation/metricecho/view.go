package metricecho

import (
	"fmt"

	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
)

func GetViews() ([]view.View, error) {
	customBucketsView, err := view.New(
		view.MatchInstrumentName("*request_duration_seconds"),
		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("meter custom view cannot set; %w", err)
	}

	return []view.View{customBucketsView}, nil
}
