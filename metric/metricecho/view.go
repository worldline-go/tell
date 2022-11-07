package metricecho

import (
	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/tglobal"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/view"
)

func GetViews() []view.View {
	customBucketView, err := view.New(
		view.MatchInstrumentName("*request_duration_seconds"),
		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
	)
	if err != nil {
		panic(err)
	}

	return []view.View{customBucketView}
}

func init() {
	tglobal.MetricViews.Add("echo", GetViews())
}
