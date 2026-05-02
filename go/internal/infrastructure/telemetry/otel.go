// Package telemetry initialises the OpenTelemetry SDK provider.
// Sprint 2 — US-008. Skeleton wired in main.go.
package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracer sets up the global OTEL TracerProvider.
// Returns a shutdown function — call it on SIGTERM.
// Supports both gRPC (localhost collector) and HTTP/protobuf (Grafana Cloud direct).
func InitTracer(ctx context.Context, serviceName, serviceNamespace, deploymentEnvironment, endpoint string) (shutdown func(context.Context) error, err error) {
	protocol := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") // "grpc" | "http/protobuf"

	var exporter sdktrace.SpanExporter

	if protocol == "http/protobuf" || strings.HasPrefix(endpoint, "https://") {
		// HTTP mode — Grafana Cloud direct (TLS + Basic Auth via OTEL_EXPORTER_OTLP_HEADERS)
		u, parseErr := url.Parse(endpoint)
		if parseErr != nil {
			return nil, fmt.Errorf("telemetry: invalid endpoint: %w", parseErr)
		}
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(u.Host),
			otlptracehttp.WithURLPath(u.Path+"/v1/traces"),
		)
	} else {
		// gRPC mode — local OTEL collector (no TLS in dev)
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("telemetry: create exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceNamespace(serviceNamespace),
			semconv.DeploymentEnvironment(deploymentEnvironment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("telemetry: create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}