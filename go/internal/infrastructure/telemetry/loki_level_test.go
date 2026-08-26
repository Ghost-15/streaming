package telemetry_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

func TestLokiWriterWriteLevelPreservesZerologLevel(t *testing.T) {
	levels := make(chan []string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Streams []struct {
				Stream map[string]string `json:"stream"`
			} `json:"streams"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		got := make([]string, 0, len(payload.Streams))
		for _, stream := range payload.Streams {
			got = append(got, stream.Stream["level"])
		}
		levels <- got
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	w, err := telemetry.NewLokiWriter(server.URL, "user", "pass", "svc", "dev")
	if err != nil {
		t.Fatalf("NewLokiWriter err = %v", err)
	}
	if _, err := w.WriteLevel(zerolog.WarnLevel, []byte(`{"message":"careful"}`)); err != nil {
		t.Fatalf("WriteLevel err = %v", err)
	}
	w.Close()

	select {
	case got := <-levels:
		if len(got) != 1 || got[0] != "warn" {
			t.Fatalf("Loki levels = %v, want [warn]", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Loki payload")
	}
}
