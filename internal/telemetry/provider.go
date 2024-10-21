package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
)

func NewProvider(serviceName, endpoint string, secure bool, protocol string, headers map[string]string) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
			semconv.DeploymentEnvironmentKey.String("development"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var exp *otlptrace.Exporter

	if protocol == "http" {
		httpOpts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
		}
		if !secure {
			httpOpts = append(httpOpts, otlptracehttp.WithInsecure())
		}
		if len(headers) > 0 {
			httpOpts = append(httpOpts, otlptracehttp.WithHeaders(headers))
		}
		exp, err = otlptracehttp.New(context.Background(), httpOpts...)
	} else {
		grpcOpts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithDialOption(grpc.WithBlock()),
		}
		if !secure {
			grpcOpts = append(grpcOpts, otlptracegrpc.WithInsecure())
		}
		if len(headers) > 0 {
			grpcOpts = append(grpcOpts, otlptracegrpc.WithHeaders(headers))
		}
		exp, err = otlptracegrpc.New(context.Background(), grpcOpts...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exp, sdktrace.WithBatchTimeout(time.Second))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)

	// Set up propagation
	prop := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	otel.SetTextMapPropagator(prop)

	return tp, nil
}
