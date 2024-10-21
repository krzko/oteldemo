package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

const (
	SeverityTrace = log.SeverityTrace
	SeverityDebug = log.SeverityDebug
	SeverityInfo  = log.SeverityInfo
	SeverityWarn  = log.SeverityWarn
	SeverityError = log.SeverityError
	SeverityFatal = log.SeverityFatal
)

func NewProvider(serviceName, endpoint string, secure bool, protocol string, headers map[string]string) (*sdktrace.TracerProvider, *sdkmetric.MeterProvider, *sdklog.LoggerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.DeploymentEnvironmentKey.String("development"),
		),
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExp, err := createTraceExporter(endpoint, secure, protocol, headers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	metricExp, err := createMetricExporter(endpoint, secure, protocol, headers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithInterval(10*time.Second))),
	)

	// Create log exporter
	var logExp sdklog.Exporter
	if protocol == "http" {
		logExp, err = createHTTPLogExporter(endpoint, secure, headers)
	} else {
		logExp, err = createGRPCLogExporter(endpoint, secure, headers)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	// Create log provider
	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		sdklog.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)

	// Set up propagation
	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(prop)

	return tp, mp, lp, nil
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

func createHTTPLogExporter(endpoint string, secure bool, headers map[string]string) (sdklog.Exporter, error) {
	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
	}
	if !secure {
		opts = append(opts, otlploghttp.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}
	return otlploghttp.New(context.Background(), opts...)
}

func createGRPCLogExporter(endpoint string, secure bool, headers map[string]string) (sdklog.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(endpoint),
		otlploggrpc.WithDialOption(grpc.WithBlock()),
	}
	if !secure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}
	if len(headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(headers))
	}
	return otlploggrpc.New(context.Background(), opts...)
}

// CreateLogRecord is a helper function to create OpenTelemetry log records
// with consistent timestamp and formatting.
func CreateLogRecord(severity log.Severity, message string, attrs ...log.KeyValue) log.Record {
	now := time.Now()
	record := log.Record{}
	record.SetTimestamp(now)
	record.SetObservedTimestamp(now)
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(log.StringValue(message))
	record.AddAttributes(attrs...)
	return record
}
