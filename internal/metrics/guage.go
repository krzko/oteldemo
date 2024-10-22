package metrics

import (
	"context"
	"math"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/trace"
)

type GaugeConfig struct {
	Name        string
	Description string
	Unit        string
	Attributes  []attribute.KeyValue
	Min         float64
	Max         float64
	Temporality metricdata.Temporality
}

func RegisterGauge(meter metric.Meter, gc GaugeConfig) (metric.Float64ObservableGauge, error) {
	gauge, err := meter.Float64ObservableGauge(
		gc.Name,
		metric.WithUnit(gc.Unit),
		metric.WithDescription(gc.Description),
	)

	if err != nil {
		return nil, err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		value := generateGaugeValue(gc.Min, gc.Max)
		attrs := gc.Attributes

		// Add trace context if available
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			traceID := span.SpanContext().TraceID().String()
			spanID := span.SpanContext().SpanID().String()
			attrs = append(attrs,
				attribute.String("trace_id", traceID),
				attribute.String("span_id", spanID),
			)
		}

		o.ObserveFloat64(gauge, value, metric.WithAttributes(attrs...))
		return nil
	}, gauge)

	return gauge, err
}

func generateGaugeValue(min, max float64) float64 {
	amplitude := (max - min) / 2
	center := min + amplitude
	return center + amplitude*math.Sin(float64(time.Now().UnixNano())/1e9)
}
