package telemetry_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestLokiWriter_NormalizesTrailingSlash(t *testing.T) {
	path := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path <- r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	w, err := telemetry.NewLokiWriter(server.URL+"/", "user", "pass", "svc", "dev")
	if err != nil {
		t.Fatalf("NewLokiWriter err = %v", err)
	}
	if _, err := w.Write([]byte(`{"level":"info","message":"hi"}`)); err != nil {
		t.Fatalf("Write err = %v", err)
	}
	defer w.Close()
	select {
	case got := <-path:
		if got != "/loki/api/v1/push" {
			t.Fatalf("request path = %q, want /loki/api/v1/push", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Loki batch")
	}
}

func TestLokiWriter_DoesNotBlockOnSlowEndpoint(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		entered <- struct{}{}
		<-release
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	w, err := telemetry.NewLokiWriter(server.URL, "user", "pass", "svc", "dev")
	if err != nil {
		t.Fatalf("NewLokiWriter err = %v", err)
	}
	started := time.Now()
	if _, err := w.Write([]byte(`{"level":"info","message":"slow"}`)); err != nil {
		t.Fatalf("Write err = %v", err)
	}
	if elapsed := time.Since(started); elapsed > 100*time.Millisecond {
		t.Fatalf("Write blocked for %s", elapsed)
	}

	select {
	case <-entered:
		close(release)
		w.Close()
	case <-time.After(2 * time.Second):
		close(release)
		w.Close()
		t.Fatal("timed out waiting for asynchronous Loki request")
	}
}

func TestLokiWriter_BatchesEntries(t *testing.T) {
	valueCount := make(chan int, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Streams []struct {
				Values [][2]string `json:"values"`
			} `json:"streams"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		count := 0
		for _, stream := range payload.Streams {
			count += len(stream.Values)
		}
		valueCount <- count
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	w, err := telemetry.NewLokiWriter(server.URL, "user", "pass", "svc", "dev")
	if err != nil {
		t.Fatalf("NewLokiWriter err = %v", err)
	}
	_, _ = w.Write([]byte(`{"message":"one"}`))
	_, _ = w.Write([]byte(`{"message":"two"}`))
	w.Close()

	if got := <-valueCount; got != 2 {
		t.Fatalf("batched values = %d, want 2", got)
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
