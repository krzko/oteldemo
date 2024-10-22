# oteldemo

`oteldemo` is a tool designed to demonstrate OpenTelemetry in action.

It simulates a social media application called "Chirper", generating synthetic telemetry data to showcase how OpenTelemetry can provide insights into complex, distributed systems in a simple, single binary.

## Features

- **Service Simulation**: Mimics a multi-tiered application including frontend, API gateway, and backend services.
- **OpenTelemetry Integration**: Demonstrates OpenTelemetry implementation in Go applications.
- **Multi-Signal Support**: 
  - **Tracing**: Generates distributed traces showing request flows.
  - **Metrics**: Produces application and system metrics.
  - **Logging**: Creates structured logs integrated with traces and metrics.

## Configuration

Configuration can be provided through environment variables or command-line flags. The following options are available:

`OTEL_EXPORTER_OTLP_ENDPOINT`: OpenTelemetry collector endpoint (default: localhost:4318)
`OTEL_EXPORTER_OTLP_SECURE`: Use secure connection (default: false)
`OTEL_EXPORTER_OTLP_PROTOCOL`: Protocol to use (grpc or http, default: http)
`APP_ENV`: Application environment (default: development)
`OTEL_EXPORTER_OTLP_HEADERS`: Headers as key=value pairs, comma-separated

## OpenTelemetry Showcase

OTelDemo highlights key OpenTelemetry features:

### Tracing
- Distributed context propagation across services
- Automatic instrumentation of HTTP clients and servers
- Manual instrumentation of business logic
- Span events for detailed operation tracking
- Span links to show relationships between operations
- Custom attributes for specific data

### Metrics
- Runtime metrics (memory usage, GC statistics)
- Request rates, durations, and error rates
- Custom business metrics (e.g., post counts, user activities)

### Logs
- Structured logging with trace correlation
- Log-injected trace IDs for easy correlation
- Various log severity levels

## Getting Started

### brew

```bash
brew install krzko/tap/oteldemo
```

## Use Cases

1. **Observability Testing**: Validate your observability stack with OpenTelemetry data.
2. **OpenTelemetry Demonstration**: Showcase OpenTelemetry's capabilities in distributed tracing.
3. **Performance Testing**: Simulate high load scenarios for observability pipeline testing.
4. **Learning**: Explore OpenTelemetry implementation in a realistic environment.
