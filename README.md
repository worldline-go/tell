# tell

This library include metric and trace helper functions to work directly in finops.

```sh
go get github.com/worldline-go/tell/metric
```

To close some metrics and trace

```sh
# if empty, metrics and trace providers and create noop provider to continue to work same as code perspective.
# default is empty
OTEL_EXPORTER_OTLP_ENDPOINT=otel-collector:4317
# Also TELEMETRY_COLLECTOR can usable for same thing
# TELEMETRY_COLLECTOR=otel-collector:4317
# inteval duration so send new metrics to otel collector (using time.Parseduration)
# default 5s
TELEMETRY_METRIC_PROVIDER_INTERVAL=5s
```

> `TELEMETRY_` prefix comes with igconfig!

## Otel Environment Values

Metric and trace checking some special environment values for collector. We should fallow to opentelemetry schemas.

Our environment already give these informations not need to do anything.

Local testing more than one service with metrics, you should also provide this informations to prevent mixing.

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

`tell.New` function also set the global values so next time when you need you can get from the `global` package.

These global get using by third-party libraries.

```go
// to get tracer provider // go.opentelemetry.io/otel
otel.GetTracerProvider()
// to get meter provider // go.opentelemetry.io/otel/metric/global
global.MeterProvider()
```

## Metric

To add some metric, use collector's MeterProvider to create a metric entry and add some values to that entry.

> Hold this meters in a struct to reach easily. Check [example](_example/telemetry/metric.go)

```go
// to get meter provider in collector
collector.MeterProvider 
// to get meter provider in global
global.MeterProvider()
```

__Counter:__

```go
successCounter, err = collector.MeterProvider.Meter("").
    Int64Counter("request_success", instrument.WithDescription("number of success count"))
if err != nil {
    log.Panic().Msgf("failed to initialize successCounter; %w", err)
}

// use counter, add attributes here to give much meaning to your counter.
successCounter.Add(c.Request().Context(), 1, attribute.Key("special").String("X"))
```

__Up/Down Counter:__ this is same as counter but it can also decrese.

```go
counterUpDown, err = collector.MeterProvider.Meter("").
    Int64UpDownCounter("request_success", instrument.WithDescription("number of success count"))
if err != nil {
    log.Panic().Msgf("failed to initialize successCounter; %w", err)
}

// use counter, add attributes here to give much meaning to your counter.
counterUpDown.Add(c.Request().Context(), 1, attribute.Key("special").String("X"))
```

__Histogram:__

```go
valuehistogram, err = collector.MeterProvider.Meter("").
    Float64Histogram("request_histogram", instrument.WithDescription("value histogram"))
if err != nil {
    log.Panic().Msgf("failed to initialize valuehistogram; %w", err)
}

// use histogram, add attributes here to give much meaning to your counter
valuehistogram.Record(c.Request().Context(), float64(countInt), attribute.Key("special").String("X"))
```

__Gauge:__ this is special and it need to be run with async and we need to register to callback. It is like background operation.

```go
meter := collector.MeterProvider.Meter("")

up, err := meter.Int64ObservableGauge("up", instrument.WithDescription("application up status"))
if err != nil {
    log.Error().Err(err).Msg("failed to set up gauge metric")
}

regUp, err := meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
    o.ObserveInt64(up, c.isUp) // value to observe
    return nil
}, up)

if err != nil {
    log.Error().Err(err).Msg("failed to register up gauge metric")
}

// shutdown will deregister
c.AddRegister(regUp)
```

### View

View is design how to looks like of your metrics. With this view you can setup your histogram bucket's values.

`Name` is important, explain which metric we will change.  
In here we used `*request_duration_seconds` because application name will come as prefix.

```go
customBucketView := metric.NewView(
    metric.Instrument{
        Name: "*request_duration_seconds",
    },
    metric.Stream{
        Aggregation: aggregation.ExplicitBucketHistogram{
            Boundaries: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
    },
)
```

If you have views, you need to add before to initialize metric provider.

Add to the tglobal

```go
tglobal.MetricViews.Add("echo", GetViews()) // accepts []view.View
```

### Example Usage

Add to the project this package [_example/telemetry/metric.go](/_example/telemetry/metric.go) to hold the custom metrics.

```go
package main

//
// config loaded
//

// open telemetry
collector, err := tell.New(ctx, cnf.Telemetry)
if err != nil {
    log.Fatal().Err(err).Msg("failed to init telemetry")
}

defer collector.Shutdown()

telemetry.AddGlobalAttr(attribute.Key("channel").String(cnf.Channel))
if err := telemetry.SetGlobalMeter(); err != nil {
    log.Fatal().Err(err).Msg("failed to set metric")
}
```

After that use your metrics

```go
// in somewhere use your metrics
telemetry.GlobalMeter.Success.Add(ctx, 1, telemetry.GlobalAttr...)
```

### Echo

#### Metric

```sh
go get github.com/worldline-go/tell/metric/metricecho
```

Use our Echo framework's middleware to share metrics.

```go
// add echo metrics
e.Use(metricecho.HTTPMetrics(nil))
```

metricecho package has init function and records views always, no need to additional settings.

#### Trace

Trace is not ready for finops, we will add details later.

```sh
go get go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho
```

```go
// add otel tracing
e.Use(otelecho.Middleware(config.LoadConfig.AppName, otelecho.WithTracerProvider(otel.GetTracerProvider())))
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

We will add data sampler in future for our finops! Currently manually handle yourself.

### Custom Trace

Start a trace with using previous context. After the start it will create new context and use that context for next trace.  
If you not use context or not give previous one tracer, it will start as root and not good for view.

```go
// collector.TracerProvider is our trace provider and it is global
// to get tracer provider // go.opentelemetry.io/otel
ctxTrace, span := otel.GetTracerProvider().Tracer(c.Path()).Start(c.Request().Context(), "PostCount")
defer span.End()

// add extra values to your trace data
// our collector adds extra fields automatically as servicename, containerid
span.SetAttributes(attribute.Key("request.count.set").Int64(countInt))
```

## Development

To test in local machine deploy otel-collector, grafana, prometheus and jaeger:

__TODO:__ export folder

```sh
cd $(mktemp -d)
curl -fksSL  | tar --overwrite -zx

docker-compose -p tell --file tell-main-compose/compose/compose.yml up -d
```

| Project                    | Port  |
|----------------------------|-------|
| grafana                    | 3000  |
| jaeger                     | 16686 |
| prometheus                 | 9090  |
| collector -> otel-grpc     | 4317  |
| collector -> otel-metric   | 8888  |
| collector -> otel-exported | 8889  |

Check the status of tell compose

```sh
docker-compose -p tell ps
```

Down the compose

```sh
docker-compose -p tell down --volumes
```

## Resources

https://github.com/open-telemetry/opentelemetry-go  
https://opentelemetry.io/registry/  
https://pkg.go.dev/go.opentelemetry.io/collector/exporter/prometheusexporter#section-readme  
https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/examples/demo  
https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md  
https://docs.docker.com/engine/swarm/services/#create-services-using-templates
