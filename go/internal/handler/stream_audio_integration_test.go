package handler_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func TestAudioIngestFanoutAndListenerCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	beforeIngest := testutil.ToFloat64(telemetry.AudioIngestBytesTotal.WithLabelValues("stream-1"))
	beforeEgress := testutil.ToFloat64(telemetry.AudioEgressBytesTotal.WithLabelValues("stream-1"))
	var listenerDelta atomic.Int32
	repo := &mock.MockStreamRepository{
		FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
			return &entity.Stream{
				ID:              "stream-1",
				BroadcasterID:   "user-1",
				Status:          entity.StreamStatusLive,
				ActiveSessionID: &testActiveSessionID,
			}, nil
		},
		IncrementListenersFn: func(_ context.Context, _ string, delta int) error {
			listenerDelta.Add(int32(delta))
			return nil
		},
	}
	hub := streaming.NewHub()
	h := handler.NewStreamHandler(
		usecase.NewStreamUseCase(repo, nil),
		handler.WithAudioStreaming(hub, time.Minute, 5*time.Second, time.Second, 1<<20, 1024, 4),
	)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("claims", &entity.JWTClaims{UserID: "user-1", Role: entity.RoleDiffuseur})
		c.Next()
	})
	engine.PUT("/streams/:id/audio", h.IngestAudio)
	engine.GET("/streams/:id/listen", h.StreamAudio)
	server := httptest.NewServer(engine)
	defer server.Close()

	audioReader, audioWriter := io.Pipe()
	ingestRequest, err := http.NewRequestWithContext(t.Context(), http.MethodPut, server.URL+"/streams/stream-1/audio", audioReader)
	if err != nil {
		t.Fatal(err)
	}
	ingestRequest.Header.Set("Content-Type", "audio/mpeg")
	ingestRequest.Header.Set("X-Stream-Session-ID", testActiveSessionID)
	ingestDone := make(chan error, 1)
	go func() {
		response, requestErr := http.DefaultClient.Do(ingestRequest)
		if response != nil {
			_ = response.Body.Close()
		}
		ingestDone <- requestErr
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, active := hub.ContentType("stream-1"); active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("publisher did not become active")
		}
		time.Sleep(5 * time.Millisecond)
	}

	listenCtx, cancelListen := context.WithCancel(context.Background())
	listenRequest, err := http.NewRequestWithContext(listenCtx, http.MethodGet, server.URL+"/streams/stream-1/listen", nil)
	if err != nil {
		t.Fatal(err)
	}
	listenResponse, err := http.DefaultClient.Do(listenRequest)
	if err != nil {
		t.Fatalf("listen request: %v", err)
	}
	defer listenResponse.Body.Close()
	if listenResponse.StatusCode != http.StatusOK {
		t.Fatalf("listen status = %d, want 200", listenResponse.StatusCode)
	}
	if got := listenResponse.Header.Get("Content-Type"); got != "audio/mpeg" {
		t.Fatalf("Content-Type = %q, want audio/mpeg", got)
	}

	payload := bytes.Repeat([]byte{0x2a}, 1024)
	if _, err := audioWriter.Write(payload); err != nil {
		t.Fatalf("write publisher audio: %v", err)
	}
	received := make([]byte, len(payload))
	if _, err := io.ReadFull(listenResponse.Body, received); err != nil {
		t.Fatalf("read listener audio: %v", err)
	}
	if !bytes.Equal(received, payload) {
		t.Fatal("listener payload differs from broadcaster payload")
	}
	if delta := testutil.ToFloat64(telemetry.AudioIngestBytesTotal.WithLabelValues("stream-1")) - beforeIngest; delta != float64(len(payload)) {
		t.Fatalf("ingest bytes delta = %.0f, want %d", delta, len(payload))
	}
	if delta := testutil.ToFloat64(telemetry.AudioEgressBytesTotal.WithLabelValues("stream-1")) - beforeEgress; delta != float64(len(payload)) {
		t.Fatalf("egress bytes delta = %.0f, want %d", delta, len(payload))
	}

	cancelListen()
	_ = audioWriter.Close()
	select {
	case err := <-ingestDone:
		if err != nil {
			t.Fatalf("ingest request: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ingestion did not finish after body close")
	}
	deadline = time.Now().Add(2 * time.Second)
	for listenerDelta.Load() != 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := listenerDelta.Load(); got != 0 {
		t.Fatalf("listener database delta = %d, want 0 after cancellation", got)
	}
}
