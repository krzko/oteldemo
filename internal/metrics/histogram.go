package metrics

import (
	"context"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HistogramConfig struct {
	Name         string
	Description  string
	Unit         string
	Attributes   []attribute.KeyValue
	Bounds       []float64
	RecordMinMax bool
}

func RegisterHistogram(meter metric.Meter, hc HistogramConfig) (metric.Float64Histogram, error) {
	histogram, err := meter.Float64Histogram(
		hc.Name,
		metric.WithDescription(hc.Description),
		metric.WithUnit(hc.Unit),
		metric.WithExplicitBucketBoundaries(hc.Bounds...),
	)
	if err != nil {
		return nil, err
	}

	return histogram, nil
}

func SimulateHistogram(ctx context.Context, histogram metric.Float64Histogram, hc HistogramConfig) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			value := generateHistogramValue(r, hc.Bounds)
			histogram.Record(ctx, value, metric.WithAttributes(hc.Attributes...))
		}
	}
}

func generateHistogramValue(r *rand.Rand, bounds []float64) float64 {
	if len(bounds) == 0 {
		return r.Float64() * 100
	}
	maxBound := bounds[len(bounds)-1]
	// Generate values with a slight bias towards lower buckets
	return r.Float64() * maxBound * 1.1
}
