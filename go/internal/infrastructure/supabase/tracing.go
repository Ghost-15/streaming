package supabase

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/Ghost-15/streaming/internal/infrastructure/supabase"

var errDatabaseUnavailable = errors.New("supabase: database unavailable")

func startRepoSpan(
	ctx context.Context,
	domain string,
	repository string,
	method string,
	table string,
	operation string,
	attrs ...attribute.KeyValue,
) (context.Context, trace.Span) {
	baseAttrs := []attribute.KeyValue{
		attribute.String("component", "repository"),
		attribute.String("repository.system", "supabase"),
		attribute.String("repository.domain", domain),
		attribute.String("repository.name", repository),
		attribute.String("repository.method", method),
		attribute.String("db.system.name", "postgresql"),
		attribute.String("db.namespace", "supabase"),
		attribute.String("db.collection.name", table),
		attribute.String("db.operation.name", operation),
	}
	baseAttrs = append(baseAttrs, attrs...)

	return otel.Tracer(tracerName).Start(
		ctx,
		"supabase."+domain+"."+method,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(baseAttrs...),
	)
}

func finishRepoSpan(span trace.Span, err *error) {
	if err != nil && *err != nil {
		span.SetAttributes(attribute.String("error.type", fmt.Sprintf("%T", *err)))
		span.SetStatus(codes.Error, "repository operation failed")
	}
	span.End()
}
