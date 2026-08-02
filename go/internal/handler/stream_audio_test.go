package handler_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

// audioEngine wires the public Audio route + the Push route (with an injected
// diffuseur claim), mirroring the real router groups.
func audioEngine(h *handler.StreamHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/streams/:id/audio", h.Audio)
	push := r.Group("/")
	push.Use(func(c *gin.Context) {
		c.Set("claims", &entity.JWTClaims{UserID: "bc-1", Role: entity.RoleDiffuseur})
		c.Next()
	})
	push.POST("/streams/:id/push", h.Push)
	return r
}

func newAudioHandler(hub *streaming.Hub) *handler.StreamHandler {
	return handler.NewStreamHandler(usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil), hub)
}

func TestStreamHandler_Push_MissingClaims(t *testing.T) {
	h := newAudioHandler(streaming.NewHub())
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/streams/:id/push", h.Push)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/streams/s1/push", bytes.NewReader([]byte("x")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Push missing claims status = %d, want 401", w.Code)
	}
}

// TestStreamingIntegration_BroadcastToListeners is the end-to-end proof:
// a broadcaster pushes an audio chunk and two connected listeners both receive it.
func TestStreamingIntegration_BroadcastToListeners(t *testing.T) {
	hub := streaming.NewHub()
	srv := httptest.NewServer(audioEngine(newAudioHandler(hub)))
	defer srv.Close()

	// Two listeners connect to the live audio stream.
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

	// Give the Hub a moment to register both clients.
	time.Sleep(150 * time.Millisecond)
	if n := hub.ListenerCount("s1"); n != 2 {
		t.Fatalf("listener count = %d, want 2", n)
	}

	// The broadcaster pushes one audio chunk.
	chunk := []byte("LIVEAUDIO")
	presp, err := http.Post(srv.URL+"/streams/s1/push", "audio/webm", bytes.NewReader(chunk))
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	presp.Body.Close()

	// Both listeners must receive exactly that chunk.
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
	srv := httptest.NewServer(audioEngine(newAudioHandler(hub)))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/streams/s1/audio")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	time.Sleep(150 * time.Millisecond)
	if hub.ListenerCount("s1") != 1 {
		t.Fatalf("listener count = %d, want 1", hub.ListenerCount("s1"))
	}

	// Disconnect and wait for the server to clean up.
	resp.Body.Close()
	deadline := time.Now().Add(2 * time.Second)
	for hub.ListenerCount("s1") != 0 && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	if n := hub.ListenerCount("s1"); n != 0 {
		t.Errorf("listener not cleaned up after disconnect: count = %d", n)
	}
}

// TestStreamHandler_Audio_LateJoinerGetsInitSegment verifies a listener that
// connects after the stream started still receives the cached init segment.
func TestStreamHandler_Audio_LateJoiner(t *testing.T) {
	hub := streaming.NewHub()
	srv := httptest.NewServer(audioEngine(newAudioHandler(hub)))
	defer srv.Close()

	// Broadcaster pushes the first chunk (cached as init segment) before anyone listens.
	initChunk := []byte("INITHEADER")
	presp, err := http.Post(srv.URL+"/streams/s1/push", "audio/webm", bytes.NewReader(initChunk))
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	presp.Body.Close()

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
