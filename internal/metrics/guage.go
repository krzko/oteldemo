package metrics

import (
	"context"
	"math"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
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

	_, err = meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		value := generateGaugeValue(gc.Min, gc.Max)
		o.ObserveFloat64(gauge, value, metric.WithAttributes(gc.Attributes...))
		return nil
	}, gauge)

	return gauge, err
}

func generateGaugeValue(min, max float64) float64 {
	amplitude := (max - min) / 2
	center := min + amplitude
	return center + amplitude*math.Sin(float64(time.Now().UnixNano())/1e9)
}
