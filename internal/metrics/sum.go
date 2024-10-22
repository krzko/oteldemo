package metrics

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
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

func RecordSum(ctx context.Context, counter metric.Int64Counter, value int64, attrs ...attribute.KeyValue) {
	opts := []metric.AddOption{metric.WithAttributes(attrs...)}

	// Add exemplar if there's an active span in the context
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()
		opts = append(opts, metric.WithAttributes(
			attribute.String("trace_id", traceID),
			attribute.String("span_id", spanID),
		))
	}

	counter.Add(ctx, value, opts...)
}
