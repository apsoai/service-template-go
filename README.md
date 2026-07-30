# Apso Go Service Template

The GitHub template behind every Apso Go backend: Gin, GORM, goose migrations, and PostgreSQL, with Swagger documentation via swaggo. `apso init --name my-api --language go` clones this repository, and `apso generate` turns your `.apsorc` schema into models, handlers, routes, and services inside it. The generated code is yours: Apache-2.0 licensed, no Apso SDK, no runtime dependency.

New to Apso? The [quickstart](https://docs.apso.ai/get-started/quickstart) goes from schema to running API in about ten minutes. All three language templates are listed at [apso.ai/templates](https://apso.ai/templates).

## Project Structure

```
.
├── cmd/
│   └── main.go              # Application entry point
├── autogen/                  # Auto-generated code (DO NOT EDIT)
│   ├── dto/                  # Data transfer objects
│   ├── handlers/             # HTTP request handlers
│   ├── models/               # Database models
│   │   └── enums/            # Enum types
│   ├── routes/               # Route registration
│   └── services/             # Business logic
├── extensions/               # Custom/manual code
│   └── middleware/           # Custom middleware
├── internal/                 # Internal (non-generated) packages
│   └── observability/        # OpenTelemetry tracing bootstrap
├── docs/                     # Generated Swagger docs
├── .apsorc                   # APSO CLI configuration
├── .env.example              # Environment variables template
├── Dockerfile                # Docker build configuration
├── go.mod                    # Go module definition
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL
- Docker (optional)

### Local Development

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Update `.env` with your database credentials

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Generate Swagger docs:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   swag init -g cmd/main.go -o docs
   ```

5. Run the server:
   ```bash
   go run cmd/main.go
   ```

### Docker

Build and run with Docker:

```bash
docker build -t my-api .
docker run -p 3000:3000 --env-file .env my-api
```

## API Documentation

Swagger documentation is available at: `http://localhost:3000/_docs`

## Observability (OpenTelemetry)

Every service built from this template ships with OpenTelemetry auto-instrumentation wired in. When enabled, each incoming HTTP request becomes a trace/span with route, method, status, and latency, and each GORM database query becomes a child span, all exported over OTLP HTTP to a collector you point it at (an OpenTelemetry Collector feeding a Grafana/Tempo stack, or Grafana Cloud's managed OTLP endpoint).

The export destination is never hardwired to an Apso endpoint. You own where telemetry goes.

### Off by default

Tracing is a no-op unless you turn it on. With no OTLP endpoint and no enable flag set, the instrumentation does not start, does not export anything, and cannot block startup on an unreachable collector. Local dev and CI run exactly as before.

### Enabling it

Set the OTLP endpoint (and optionally a service name):

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
export OTEL_SERVICE_NAME=my-api
go run cmd/main.go
```

You should see `[otel] tracing enabled (...)` on startup, and traces begin flowing to your collector. Buffered spans are flushed on `SIGINT`/`SIGTERM` before the process exits.

### Environment variables

| Variable | Description | Default |
| --- | --- | --- |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP HTTP endpoint (OpenTelemetry Collector or Grafana Cloud). Setting this enables tracing. | unset (tracing off) |
| `OTEL_SERVICE_NAME` | `service.name` reported on every trace. | falls back to `APP_NAME`, then the `.apsorc` `serviceName`, then `apso-go-service` |
| `OTEL_ENABLED` | Force-enable (`true`/`1`/`yes`/`on`) or force-disable (anything else), overriding endpoint detection. | unset |

Standard `OTEL_*` environment variables (headers, resource attributes, sampling) are honored by the OpenTelemetry SDK as well.

### How it is wired

`internal/observability/tracing.go` builds an OTLP HTTP trace exporter and a `TracerProvider` with `service.name` / `service.version` resource attributes, and registers it as the global provider. In `cmd/main.go` the Gin engine gets `otelgin.Middleware(serviceName)` and the GORM DB gets the OpenTelemetry tracing plugin, both only when tracing is enabled so the default request and query paths are untouched when it is off. The provider is shut down (flushing in-flight spans) as part of the server's graceful shutdown.

## APSO CLI Integration

This template is designed to work with the APSO CLI. To add entities:

```bash
apso scaffold
```

The CLI will generate:
- Models in `autogen/models/`
- DTOs in `autogen/dto/`
- Handlers in `autogen/handlers/`
- Services in `autogen/services/`
- Routes in `autogen/routes/`

## Extending the Template

Add custom code in the `extensions/` folder:
- Custom middleware in `extensions/middleware/`
- Create additional folders as needed for your custom logic

**Important:** Never modify files in the `autogen/` folder as they will be overwritten by the APSO CLI.

## License

Apache License 2.0. See the [LICENSE](LICENSE) file.
