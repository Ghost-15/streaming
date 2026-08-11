// Package telemetry initialises the OpenTelemetry SDK provider.
// Sprint 2 — US-008. Skeleton wired in main.go.
package telemetry

import (
	"context"
	"fmt"
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
		u, parseErr := parseOTLPHTTPEndpoint(endpoint, "traces")
		if parseErr != nil {
			return nil, parseErr
		}
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(u.Host),
			otlptracehttp.WithURLPath(u.Path),
		}
		if u.Scheme == "http" {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		exporter, err = otlptracehttp.New(ctx, opts...)
	} else {
		// gRPC mode — local OTEL collector (no TLS in dev)
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(trimEndpointScheme(endpoint)),
			otlptracegrpc.WithInsecure(),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("telemetry: create exporter: %w", err)
	}

	res, err := newResource(ctx, serviceName, serviceNamespace, deploymentEnvironment)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func newResource(ctx context.Context, serviceName, serviceNamespace, deploymentEnvironment string) (*resource.Resource, error) {
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
	return res, nil
}
