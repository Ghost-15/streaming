package telemetry_test

import (
	"context"
	"testing"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

func TestNewLokiWriter_MissingConfig(t *testing.T) {
	if _, err := telemetry.NewLokiWriter("", "", "", "svc", "dev"); err == nil {
		t.Error("NewLokiWriter with empty config: expected error")
	}
}

func TestLokiWriter_WriteNonBlocking(t *testing.T) {
	w, err := telemetry.NewLokiWriter("http://127.0.0.1:1/nope", "user", "pass", "svc", "dev")
	if err != nil {
		t.Fatalf("NewLokiWriter err = %v", err)
	}
	defer w.Close()

	// The Loki endpoint is unreachable; Write must not block or error (best-effort).
	n, err := w.Write([]byte(`{"level":"info","message":"hi"}`))
	if err != nil {
		t.Errorf("Write err = %v, want nil", err)
	}
	if n == 0 {
		t.Error("Write returned 0 bytes")
	}
}

func TestInitTracer_GRPC(t *testing.T) {
	shutdown, err := telemetry.InitTracer(context.Background(), "svc", "ns", "dev", "localhost:4317")
	if err != nil {
		t.Fatalf("InitTracer grpc err = %v", err)
	}
	if shutdown == nil {
		t.Fatal("InitTracer returned nil shutdown")
	}
	if err := shutdown(context.Background()); err != nil {
		t.Errorf("shutdown err = %v", err)
	}
}

func TestInitTracer_HTTP(t *testing.T) {
	shutdown, err := telemetry.InitTracer(context.Background(), "svc", "ns", "dev", "https://example.com/otlp")
	if err != nil {
		t.Fatalf("InitTracer http err = %v", err)
	}
	_ = shutdown(context.Background())
}
