package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

func NewProvider(serviceName, endpoint string, secure bool, protocol string, headers map[string]string) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.DeploymentEnvironmentKey.String("development"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExp, err := createTraceExporter(endpoint, secure, protocol, headers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	metricExp, err := createMetricExporter(endpoint, secure, protocol, headers)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp, sdktrace.WithBatchTimeout(time.Second))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(10*time.Second))),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	// Set up propagation
	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(prop)

	return tp, mp, nil
}

func createTraceExporter(endpoint string, secure bool, protocol string, headers map[string]string) (*otlptrace.Exporter, error) {
	if protocol == "http" {
		return createHTTPTraceExporter(endpoint, secure, headers)
	}
	return createGRPCTraceExporter(endpoint, secure, headers)
}

func createHTTPTraceExporter(endpoint string, secure bool, headers map[string]string) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
	}
	if !secure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}
	return otlptracehttp.New(context.Background(), opts...)
}

func createGRPCTraceExporter(endpoint string, secure bool, headers map[string]string) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	}
	if !secure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlptracegrpc.WithHeaders(headers))
	}
	return otlptracegrpc.New(context.Background(), opts...)
}

func createMetricExporter(endpoint string, secure bool, protocol string, headers map[string]string) (sdkmetric.Exporter, error) {
	if protocol == "http" {
		return createHTTPMetricExporter(endpoint, secure, headers)
	}
	return createGRPCMetricExporter(endpoint, secure, headers)
}

func createHTTPMetricExporter(endpoint string, secure bool, headers map[string]string) (sdkmetric.Exporter, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(endpoint),
	}
	if !secure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(headers))
	}
	return otlpmetrichttp.New(context.Background(), opts...)
}

func createGRPCMetricExporter(endpoint string, secure bool, headers map[string]string) (sdkmetric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(endpoint),
		otlpmetricgrpc.WithDialOption(grpc.WithBlock()),
	}
	if !secure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlpmetricgrpc.WithHeaders(headers))
	}
	return otlpmetricgrpc.New(context.Background(), opts...)
}
