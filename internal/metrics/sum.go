package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type SumConfig struct {
	Name        string
	Description string
	Unit        string
	Attributes  []attribute.KeyValue
	IsMonotonic bool
}

func RegisterSum(meter metric.Meter, sc SumConfig) (metric.Int64Counter, error) {
	counter, err := meter.Int64Counter(
		sc.Name,
		metric.WithDescription(sc.Description),
		metric.WithUnit(sc.Unit),
	)
	if err != nil {
		return nil, err
	}

	return counter, nil
}

func RecordSum(ctx context.Context, counter metric.Int64Counter, value int64, attributes ...attribute.KeyValue) {
	counter.Add(ctx, value, metric.WithAttributes(attributes...))
}
