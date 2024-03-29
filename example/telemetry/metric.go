package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	GlobalAttr  []attribute.KeyValue
	GlobalMeter *Meter
)

type Meter struct {
	Error     metric.Int64Counter
	Processed metric.Int64Counter
	Rules     metric.Int64Counter
}

func AddGlobalAttr(v ...attribute.KeyValue) {
	GlobalAttr = append(GlobalAttr, v...)
}

func ExtendGlobalAttr(v ...attribute.KeyValue) []attribute.KeyValue {
	return append(GlobalAttr, v...)
}

func SetGlobalMeter() error {
	mp := otel.GetMeterProvider()

	m := &Meter{}

	var err error

	meter := mp.Meter("")

	//nolint:lll // description
	m.Processed, err = meter.Int64Counter("transaction_validator_processed_total", metric.WithDescription("number of successfully validated count"))
	if err != nil {
		return fmt.Errorf("failed to initialize transaction_validator_processed_total; %w", err)
	}

	m.Error, err = meter.Int64Counter("transaction_validator_error_total", metric.WithDescription("number of error on validation count"))
	if err != nil {
		return fmt.Errorf("failed to initialize transaction_validator_error_total; %w", err)
	}

	m.Rules, err = meter.Int64Counter("transaction_validator_rules_total", metric.WithDescription("number of used rule on validation count"))
	if err != nil {
		return fmt.Errorf("failed to initialize transaction_validator_error_total; %w", err)
	}

	GlobalMeter = m

	return nil
}

//nolint:gochecknoinits // set noop
func init() {
	_ = SetGlobalMeter()
}
