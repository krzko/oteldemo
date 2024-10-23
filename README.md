# oteldemo

`oteldemo` is a tool designed to demonstrate OpenTelemetry in action through a lightweight, single binary approach.

## About

Whilst the [OpenTelemetry Demo Application (Astronomy Shop)](https://github.com/open-telemetry/opentelemetry-demo) provides a comprehensive demonstration of OpenTelemetry features, it requires significant resources to run - including Kubernetes clusters and multiple Docker containers. `oteldemo` takes a different approach by providing a single, lightweight binary that demonstrates all the key features of OpenTelemetry instrumentation signals without the operational overhead.

It simulates a social media application called "Chirper", generating synthetic telemetry data to showcase how OpenTelemetry can provide insights into complex, distributed systems. The simulation creates realistic patterns of traces, metrics, and logs that you would typically see in a production microservices environment, but without the need to deploy actual services.

## Features

- **Lightweight**: Single binary, minimal resource requirements, no external dependencies
- **Service Simulation**: Mimics a multi-tiered application including frontend, API gateway, and backend services
- **OpenTelemetry Integration**: Demonstrates OpenTelemetry implementation in Go applications
- **Multi-Signal Support**: 
  - **Tracing**: Generates distributed traces showing request flows
  - **Metrics**: Produces application and system metrics
  - **Logging**: Creates structured logs integrated with traces and metrics

## Getting Started

Install `oteldemo` using `brew or download the binary from the [releases page](https://github.com/krzko/oteldemo/releases), you can also run it with `docker`:

### Install using brew

```bash
brew install krzko/tap/oteldemo
```

### Run it with Docker

```bash
docker run -it --rm ghcr.io/krzko/oteldemo:latest -v
```
### Run

Run the `oteldemo` binary with the desired configuration:

```bash
# Local collector with HTTP:
oteldemo -endpoint="localhost:4318" -protocol="http" -secure="false"

# Remote collector with gRPC and authentication:
oteldemo -endpoint="api.domain.tld:443" -headers="x-api-key=xxx,x-team=xxx" -protocol="grpc" -secure="true"

# Using environment variables:
export OTEL_EXPORTER_OTLP_ENDPOINT="localhost:4318"
export OTEL_EXPORTER_OTLP_PROTOCOL="http"
oteldemo
```

## Configuration

Flags can be set via command line or environment variables. Command line flags take precedence over environment variables.

| Flag | Env | Description | Default |
|------|-----|-------------|---------|
| `-endpoint` | `OTEL_EXPORTER_OTLP_ENDPOINT` | OpenTelemetry collector endpoint | `localhost:4318` |
| `-protocol` | `OTEL_EXPORTER_OTLP_PROTOCOL` | Protocol to use (grpc or http) | `http` |
| `-secure` | `OTEL_EXPORTER_OTLP_SECURE` | Use secure connection | `false` |
| `-headers` | `OTEL_EXPORTER_OTLP_HEADERS` | Headers as key=value pairs, comma-separated | |
| `-env` | `APP_ENV` | Application environment | `development` |
| `-v`, `-version` |  | Application version | |

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

## Use Cases

1. **Observability Testing**: Validate your observability stack with OpenTelemetry data.
2. **OpenTelemetry Demonstration**: Showcase OpenTelemetry's capabilities in distributed tracing.
3. **Performance Testing**: Simulate high load scenarios for observability pipeline testing.
4. **Learning**: Explore OpenTelemetry implementation in a realistic environment.
