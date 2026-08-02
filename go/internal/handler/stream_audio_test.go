package handler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// publicAudioEngine wires the public GET /audio listener route, mirroring the
// real router group (browser <audio> connects here without authentication).
func publicAudioEngine(h *handler.StreamHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/streams/:id/audio", h.Audio)
	return r
}

func newPublicAudioHandler(hub *streaming.Hub) *handler.StreamHandler {
	return handler.NewStreamHandler(
		usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil),
		handler.WithAudioStreaming(hub, time.Minute, 5*time.Second, time.Second, 1<<20, 1024, 64),
	)
}

// TestStreamingIntegration_BroadcastToListeners is the end-to-end proof on the
// public audio path: two listeners connect and both receive a broadcast chunk.
func TestStreamingIntegration_BroadcastToListeners(t *testing.T) {
	hub := streaming.NewHub()
	srv := httptest.NewServer(publicAudioEngine(newPublicAudioHandler(hub)))
	defer srv.Close()

	resp1, err := http.Get(srv.URL + "/streams/s1/audio")
	if err != nil {
		t.Fatalf("listener 1 connect: %v", err)
	}
	defer resp1.Body.Close()
	resp2, err := http.Get(srv.URL + "/streams/s1/audio")
	if err != nil {
		t.Fatalf("listener 2 connect: %v", err)
	}
	defer resp2.Body.Close()

	// Wait for both clients to register in the Hub.
	deadline := time.Now().Add(2 * time.Second)
	for hub.ListenerCount("s1") != 2 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if n := hub.ListenerCount("s1"); n != 2 {
		t.Fatalf("listener count = %d, want 2", n)
	}

	// The broadcaster fans out one audio chunk through the Hub.
	chunk := []byte("LIVEAUDIO")
	hub.Broadcast("s1", chunk)

	for i, resp := range []*http.Response{resp1, resp2} {
		buf := make([]byte, len(chunk))
		done := make(chan error, 1)
		go func(b []byte, body io.Reader) { _, e := io.ReadFull(body, b); done <- e }(buf, resp.Body)
		select {
		case e := <-done:
			if e != nil {
				t.Fatalf("listener %d read: %v", i+1, e)
			}
			if string(buf) != string(chunk) {
				t.Fatalf("listener %d got %q, want %q", i+1, buf, chunk)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("listener %d timed out waiting for audio", i+1)
		}
	}
}

// TestStreamingIntegration_Disconnect verifies the Hub releases a listener
// (goroutine + channel) when the client disconnects.
func TestStreamingIntegration_Disconnect(t *testing.T) {
	hub := streaming.NewHub()
	srv := httptest.NewServer(publicAudioEngine(newPublicAudioHandler(hub)))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/streams/s1/audio")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for hub.ListenerCount("s1") != 1 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if hub.ListenerCount("s1") != 1 {
		t.Fatalf("listener count = %d, want 1", hub.ListenerCount("s1"))
	}

	// Disconnect and wait for the server to clean up.
	resp.Body.Close()
	deadline = time.Now().Add(2 * time.Second)
	for hub.ListenerCount("s1") != 0 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if n := hub.ListenerCount("s1"); n != 0 {
		t.Errorf("listener not cleaned up after disconnect: count = %d", n)
	}
}

// TestStreamHandler_Audio_LateJoiner verifies a listener that connects after the
// stream started still receives the cached init segment.
func TestStreamHandler_Audio_LateJoiner(t *testing.T) {
	hub := streaming.NewHub()
	srv := httptest.NewServer(publicAudioEngine(newPublicAudioHandler(hub)))
	defer srv.Close()

	// The broadcaster's first chunk is cached as the init segment before anyone listens.
	initChunk := []byte("INITHEADER")
	hub.SetInitSegment("s1", initChunk)

	// Late joiner connects and should immediately get the cached init segment.
	resp, err := http.Get(srv.URL + "/streams/s1/audio")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer resp.Body.Close()

	buf := make([]byte, len(initChunk))
	done := make(chan error, 1)
	go func() { _, e := io.ReadFull(resp.Body, buf); done <- e }()
	select {
	case e := <-done:
		if e != nil {
			t.Fatalf("read init: %v", e)
		}
		if string(buf) != string(initChunk) {
			t.Fatalf("late joiner got %q, want init %q", buf, initChunk)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("late joiner timed out waiting for init segment")
	}
}
