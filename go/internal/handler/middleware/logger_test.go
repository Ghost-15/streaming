package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/trace"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
)

func TestZerologMiddleware_IncludesTraceID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buffer bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&buffer).With().Timestamp().Logger()
	t.Cleanup(func() {
		log.Logger = previousLogger
	})

	router := gin.New()
	router.Use(middleware.ZerologMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	traceID := trace.TraceID{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe}
	spanID := trace.SpanID{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ping", nil)
	req = req.WithContext(trace.ContextWithSpanContext(req.Context(), spanContext))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNoContent)
	}

	output := strings.TrimSpace(buffer.String())
	if output == "" {
		t.Fatal("expected a JSON log line, got empty output")
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		t.Fatalf("log output is not valid JSON: %v\n%s", err, output)
	}

	if got, want := payload["trace_id"], traceID.String(); got != want {
		t.Fatalf("unexpected trace_id: got %v want %v", got, want)
	}
	if got, want := payload["route"], "/ping"; got != want {
		t.Fatalf("unexpected route: got %v want %v", got, want)
	}
	if got, want := payload["status"], float64(http.StatusNoContent); got != want {
		t.Fatalf("unexpected status: got %v want %v", got, want)
	}
	if _, ok := payload["user_id"]; !ok {
		t.Fatal("expected user_id field to be present in the log payload")
	}
}

func TestRequestScopedLogger_PropagatesTraceIDToHandlerLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buffer bytes.Buffer
	previousLogger := log.Logger
	log.Logger = zerolog.New(&buffer).With().Timestamp().Logger()
	t.Cleanup(func() {
		log.Logger = previousLogger
	})

	router := gin.New()
	router.Use(middleware.ZerologMiddleware())
	router.GET("/ping", func(c *gin.Context) {
		middleware.Logger(c).Info().Msg("inside handler")
		c.Status(http.StatusNoContent)
	})

	traceID := trace.TraceID{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99}
	spanID := trace.SpanID{0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22}
	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ping", nil)
	req = req.WithContext(trace.ContextWithSpanContext(req.Context(), spanContext))

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNoContent)
	}

	lines := strings.Split(strings.TrimSpace(buffer.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected at least two log lines, got %d: %q", len(lines), buffer.String())
	}

	handlerLogFound := false
	for _, line := range lines {
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("log output is not valid JSON: %v\n%s", err, line)
		}
		if got, want := payload["trace_id"], traceID.String(); got != want {
			t.Fatalf("unexpected trace_id: got %v want %v", got, want)
		}
		if got, ok := payload["message"].(string); ok && got == "inside handler" {
			handlerLogFound = true
		}
	}

	if !handlerLogFound {
		t.Fatal("expected a handler log line with message 'inside handler'")
	}
}

// TestLokiWriter_UsesGlobalLogger verifies that the middleware uses the global
// zerolog.Logger — meaning any MultiLevelWriter (stdout + Loki) set in main.go
// is automatically used by all handler logs without extra configuration.
func TestLokiWriter_UsesGlobalLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var buf bytes.Buffer
	previousLogger := log.Logger
	// Simulate main.go wiring: replace global logger with a custom writer (here a buffer).
	log.Logger = zerolog.New(&buf).With().Timestamp().Logger()
	t.Cleanup(func() { log.Logger = previousLogger })

	router := gin.New()
	router.Use(middleware.ZerologMiddleware())
	router.POST("/api/v1/auth/login", func(c *gin.Context) {
		middleware.Logger(c).Info().Msg("user authenticated")
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/auth/login", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Fatal("expected log output, got empty — global logger not used")
	}

	// Every line must be valid JSON and contain trace_id.
	for _, line := range strings.Split(output, "\n") {
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			t.Fatalf("log line is not valid JSON: %v\n%s", err, line)
		}
		if _, ok := payload["trace_id"]; !ok {
			t.Fatalf("missing trace_id in log line: %s", line)
		}
	}
}
