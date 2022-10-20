# tell

This library include metric and trace helper functions to work directly in finops.

```sh
go get gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell
```

## Environment Values

Metric and trace checking some special environment values for collector. We should fallow to opentelemetry schemas.

```sh
# OTEL_SERVICE_NAME=transaction_api
OTEL_RESOURCE_ATTRIBUTES=service.name=transaction_api,service.instance.id=xyz123
```

In our stack this is show like that for swarm:

```env
OTEL_RESOURCE_ATTRIBUTES=service.name={{.Service.Name}},service.instance.id={{.Task.ID}},host.id={{.Node.ID}},host.name={{.Node.Hostname}}
```

Check much more details of attributes in here [opentelemetry-specification](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md)

You can also add your own values after that, these are global values for all services need to have.

## Initialize

Add this configuration in your application config struct:

```go
type Config struct {
    // Telemetry configurations
    Telemetry tell.Config
}
```

`igconfig` can handle our default values, don't need to change configuration.

After that in main of program pass the telemetry config to create new collector which is connection collector and initialize telemetry and trace providers with common attributes.

```go
collector, err := tell.New(ctx, cfg.Telemetry)
if err != nil {
    return fmt.Errorf("failed to init telemetry; %w", err)
}

defer collector.Shutdown()
```

Now you initialized and connected to our collector. You can send some metrics and trace data.

## Metric

To add some metric, use collector's MeterProvider to create a metric entry and add some values to that entry.

Hold this meters in a struct to reach easily.

__Counter:__

```go
metrics.successCounter, err = h.Meter.Meter("").
    SyncInt64().Counter("request_success", instrument.WithDescription("number of success count"))
if err != nil {
    log.Panic().Msgf("failed to initialize successCounter; %w", err)
}

// use counter, add attributes here to give much meaning to your counter.
metrics.successCounter.Add(c.Request().Context(), 1, attribute.Key("special").String("X"))
```

__Up/Down Counter:__ this is same as counter but it can also decrese.

```go
metrics.successCounter, err = h.Meter.Meter("").
    SyncInt64().UpDownCounter("request_success", instrument.WithDescription("number of success count"))
if err != nil {
    log.Panic().Msgf("failed to initialize successCounter; %w", err)
}

// use counter, add attributes here to give much meaning to your counter.
metrics.counterUpDown.Add(c.Request().Context(), 1, attribute.Key("special").String("X"))
```

__Histogram:__

```go
metrics.valuehistogram, err = h.Meter.Meter("").
    SyncFloat64().Histogram("request_histogram", instrument.WithDescription("value histogram"))
if err != nil {
    log.Panic().Msgf("failed to initialize valuehistogram; %w", err)
}

// use histogram, add attributes here to give much meaning to your counter
metrics.valuehistogram.Record(c.Request().Context(), float64(countInt), attribute.Key("special").String("X"))
```

__Gauge:__ this is special and it need to be run with async and we need to register to callback. It is like background operation.

```go
metrics.sendGauge, err = h.Meter.Meter("").AsyncInt64().Gauge("send", instrument.WithDescription("async gauge"))
if err != nil {
    log.Panic().Msgf("failed to initialize sendGauge; %w", err)
}

// check value will be checked in this callback
h.Meter.Meter("").RegisterCallback([]instrument.Asynchronous{h.metrics.sendGauge}, func(ctx context.Context) {
    h.metrics.sendGauge.Observe(ctx, checkValue, attribute.Key("special").String("X"))
})
```

### View

View is design how to looks like of your metrics. With this view you can setup your histogram bucket's values.

`MatchInstrumentName` is important, explain which metric we will change.  
In here we used `*request_duration_seconds` because application name will come as prefix.

```go
customBucketView, err := view.New(
		view.MatchInstrumentName("*request_duration_seconds"),
		view.WithSetAggregation(aggregation.ExplicitBucketHistogram{
			Boundaries: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("meter custom view cannot set; %w", err)
	}
```

After that when initializing the collector, pass this views.

```go
collector, err := tell.New(ctx, cfg.Telemetry, customBucketView)
```

### Echo

```sh
go get gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell/metric/instrumentation/metricecho
```

Use our Echo framework's middleware to share metrics.

Before to generate collector, add the echo's views

```go
metricEchoViews, err := metricecho.GetViews()
if err != nil {
	return fmt.Errorf("failed to get metricecho views; %w", err)
}

collector, err := tell.New(ctx, cfg.Telemetry, metricEchoViews...)
```

After that just enable middleware of metricecho

```go
// add echo metrics
e.Use(metricecho.HTTPMetrics(nil))
```

### Runtime

```sh
go get go.opentelemetry.io/contrib/instrumentation/runtime
```

```go
if err := runtime.Start(); err != nil {
    return fmt.Errorf("failed to start runtime metrics; %w", err)
}
```

### Others

Check the open telemetry's registry page, new instruments can add here.

https://opentelemetry.io/registry/?language=go

## Trace

Create a collector, also our collector will create trace provider.

```go
collector, err := tell.New(ctx, cfg.Telemetry)
```

Use to trace provider to create some trace data.

We will add provider data sampler in future for our finops!
