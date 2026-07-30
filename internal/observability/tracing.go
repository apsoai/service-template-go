// Package observability provides built-in OpenTelemetry tracing for the service.
//
// Tracing is a no-op by default. It only starts when explicitly enabled (see
// Enabled), so local development and CI behave exactly as before and startup is
// never blocked on an unreachable collector.
package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	// defaultServiceName is used when no service name can be resolved from the
	// environment or the .apsorc file.
	defaultServiceName = "apso-go-service"
	// serviceVersion is reported as service.version on every trace.
	serviceVersion = "1.0.0"
)

// noopShutdown is returned whenever tracing is disabled so callers can always
// defer the shutdown func without a nil check.
func noopShutdown(context.Context) error { return nil }

// Enabled reports whether tracing should be initialized.
//
// Precedence mirrors the TS template:
//   - OTEL_ENABLED, when set, is authoritative (truthy enables, falsy disables).
//   - Otherwise, tracing is enabled only when OTEL_EXPORTER_OTLP_ENDPOINT is set.
func Enabled() bool {
	if v, ok := os.LookupEnv("OTEL_ENABLED"); ok {
		return truthy(v)
	}
	return strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")) != ""
}

// ServiceName resolves the service.name reported on traces:
// OTEL_SERVICE_NAME, then APP_NAME, then the .apsorc serviceName (if readable),
// then a sensible default.
func ServiceName() string {
	if v := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("APP_NAME")); v != "" {
		return v
	}
	if v := serviceNameFromApsorc(); v != "" {
		return v
	}
	return defaultServiceName
}

// InitTracing sets up an OTLP HTTP trace exporter and a global TracerProvider.
//
// When tracing is disabled it is a pure no-op: no exporter is registered, no
// network connection is attempted, and the returned shutdown func does nothing.
// InitTracing never panics; setup failures are returned as errors so the caller
// can decide to continue without tracing rather than crashing the service.
func InitTracing(ctx context.Context) (shutdown func(context.Context) error, err error) {
	if !Enabled() {
		return noopShutdown, nil
	}

	// The OTLP HTTP exporter reads OTEL_EXPORTER_OTLP_ENDPOINT (and related
	// OTEL_* vars) from the environment. Construction does not dial the
	// collector, so an unreachable endpoint cannot block startup here.
	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return noopShutdown, fmt.Errorf("otel: create OTLP trace exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName()),
			semconv.ServiceVersion(serviceVersion),
		),
	)
	if err != nil {
		// Fall back to a resource carrying just our attributes rather than failing.
		res = resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(ServiceName()),
			semconv.ServiceVersion(serviceVersion),
		)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	fmt.Printf("[otel] tracing enabled (service.name=%s, endpoint=%s)\n",
		ServiceName(), endpointForLog())

	return tp.Shutdown, nil
}

// truthy interprets common truthy string values.
func truthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// endpointForLog returns the configured endpoint for logging purposes only.
func endpointForLog() string {
	if v := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")); v != "" {
		return v
	}
	return "(SDK default)"
}

// serviceNameFromApsorc reads the serviceName field from .apsorc, if present.
// Any error (missing file, bad JSON) is swallowed and returns an empty string.
func serviceNameFromApsorc() string {
	for _, path := range []string{".apsorc", "../.apsorc"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var cfg struct {
			ServiceName string `json:"serviceName"`
			Name        string `json:"name"`
		}
		if err := json.Unmarshal(data, &cfg); err != nil {
			continue
		}
		if strings.TrimSpace(cfg.ServiceName) != "" {
			return cfg.ServiceName
		}
		if strings.TrimSpace(cfg.Name) != "" {
			return cfg.Name
		}
	}
	return ""
}
