package telemetry

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
)

var (
	GlobalAttr  []attribute.KeyValue
	GlobalMeter *Meter
)

type Meter struct {
	Error     syncint64.Counter
	Processed syncint64.Counter
	Rules     syncint64.Counter
}

func AddGlobalAttr(v ...attribute.KeyValue) {
	GlobalAttr = append(GlobalAttr, v...)
}

func ExtendGlobalAttr(v ...attribute.KeyValue) []attribute.KeyValue {
	return append(GlobalAttr, v...)
}

func SetGlobalMeter() error {
	mp := global.MeterProvider()

	m := &Meter{}

	var err error

	//nolint:lll // description
	m.Processed, err = mp.Meter("").SyncInt64().Counter("transaction_validator_processed_total", instrument.WithDescription("number of successfully validated count"))
	if err != nil {
		return fmt.Errorf("failed to initialize transaction_validator_processed_total; %w", err)
	}

	//nolint:lll // description
	m.Error, err = mp.Meter("").SyncInt64().Counter("transaction_validator_error_total", instrument.WithDescription("number of error on validation count"))
	if err != nil {
		return fmt.Errorf("failed to initialize transaction_validator_error_total; %w", err)
	}

	//nolint:lll // description
	m.Rules, err = mp.Meter("").SyncInt64().Counter("transaction_validator_rules_total", instrument.WithDescription("number of used rule on validation count"))
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
