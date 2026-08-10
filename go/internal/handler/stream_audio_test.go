package handler_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

var testWebMClusterID = []byte{0x1f, 0x43, 0xb6, 0x75}

func testWebMCluster(payload string) []byte {
	packet := append([]byte(nil), testWebMClusterID...)
	return append(packet, []byte(payload)...)
}

// publicAudioEngine wires the public GET /audio listener route, mirroring the
// real router group (browser <audio> connects here without authentication).
func publicAudioEngine(h *handler.StreamHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/streams/:id/audio", h.Audio)
	r.GET("/streams/:id/audio/ws", h.AudioSocket)
	return r
}

func TestStreamingWebSocket_DeliversInitializationThenCluster(t *testing.T) {
	hub := streaming.NewHub()
	if err := hub.OpenChunkPublisher("s1", "audio/webm; codecs=opus"); err != nil {
		t.Fatal(err)
	}
	initChunk := []byte("WEBM-INIT-BLOB")
	hub.SetInitSegment("s1", initChunk)
	srv := httptest.NewServer(publicAudioEngine(newPublicAudioHandler(hub)))
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/streams/s1/audio/ws"
	conn, err := websocket.Dial(wsURL, "", srv.URL)
	if err != nil {
		t.Fatalf("websocket connect: %v", err)
	}
	defer conn.Close()

	var gotInit []byte
	if err := websocket.Message.Receive(conn, &gotInit); err != nil {
		t.Fatalf("receive init blob: %v", err)
	}
	if !bytes.Equal(gotInit, initChunk) {
		t.Fatalf("init blob = %q, want %q", gotInit, initChunk)
	}

	mediaChunk := testWebMCluster("ONE-COMPLETE-WEBM-CLUSTER")
	hub.Broadcast("s1", mediaChunk)
	var gotMedia []byte
	if err := websocket.Message.Receive(conn, &gotMedia); err != nil {
		t.Fatalf("receive media blob: %v", err)
	}
	if !bytes.Equal(gotMedia, mediaChunk) {
		t.Fatalf("media blob = %q, want %q", gotMedia, mediaChunk)
	}
}

func TestStreamingWebSocket_LateListenerResumesAtNextCluster(t *testing.T) {
	hub := streaming.NewHub()
	if err := hub.OpenChunkPublisher("s1", "audio/webm; codecs=opus"); err != nil {
		t.Fatal(err)
	}
	hub.SetInitSegment("s1", []byte("WEBM-METADATA"))
	srv := httptest.NewServer(publicAudioEngine(newPublicAudioHandler(hub)))
	defer srv.Close()

	wsURL := "ws" + srv.URL[len("http"):] + "/streams/s1/audio/ws"
	conn, err := websocket.Dial(wsURL, "", srv.URL)
	if err != nil {
		t.Fatalf("websocket connect: %v", err)
	}
	defer conn.Close()

	var initialization []byte
	if err := websocket.Message.Receive(conn, &initialization); err != nil {
		t.Fatalf("receive initialization: %v", err)
	}

	type received struct {
		data []byte
		err  error
	}
	receivedMedia := make(chan received, 1)
	go func() {
		var data []byte
		err := websocket.Message.Receive(conn, &data)
		receivedMedia <- received{data: data, err: err}
	}()

	hub.Broadcast("s1", []byte("continuation-without-cluster"))
	select {
	case got := <-receivedMedia:
		t.Fatalf("received unaligned media %x (err=%v)", got.data, got.err)
	case <-time.After(100 * time.Millisecond):
	}

	want := testWebMCluster("CURRENT-AUDIO")
	hub.Broadcast("s1", want)
	select {
	case got := <-receivedMedia:
		if got.err != nil {
			t.Fatalf("receive aligned media: %v", got.err)
		}
		if !bytes.Equal(got.data, want) {
			t.Fatalf("aligned media = %x, want %x", got.data, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for aligned WebM Cluster")
	}
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
	if err := hub.OpenChunkPublisher("s1", "audio/webm; codecs=opus"); err != nil {
		t.Fatal(err)
	}
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

func TestStreamingIntegration_IdleTimeoutEndsWithCleanHTTPBody(t *testing.T) {
	hub := streaming.NewHub()
	if err := hub.OpenChunkPublisher("s1", "audio/mpeg"); err != nil {
		t.Fatal(err)
	}
	h := handler.NewStreamHandler(
		usecase.NewStreamUseCase(&mock.MockStreamRepository{}, nil),
		handler.WithAudioStreaming(hub, time.Minute, 80*time.Millisecond, 10*time.Millisecond, 1<<20, 1024, 64),
	)
	srv := httptest.NewServer(publicAudioEngine(h))
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/streams/s1/audio")
	if err != nil {
		t.Fatalf("listener connect: %v", err)
	}
	defer resp.Body.Close()

	deadline := time.Now().Add(time.Second)
	for hub.ListenerCount("s1") != 1 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	hub.Broadcast("s1", []byte("LIVEAUDIO"))

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body after idle timeout: %v", err)
	}
	if string(body) != "LIVEAUDIO" {
		t.Fatalf("body = %q, want LIVEAUDIO", body)
	}
}

// TestStreamingIntegration_Disconnect verifies the Hub releases a listener
// (goroutine + channel) when the client disconnects.
func TestStreamingIntegration_Disconnect(t *testing.T) {
	hub := streaming.NewHub()
	if err := hub.OpenChunkPublisher("s1", "audio/webm; codecs=opus"); err != nil {
		t.Fatal(err)
	}
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
	if err := hub.OpenChunkPublisher("s1", "audio/webm; codecs=opus"); err != nil {
		t.Fatal(err)
	}
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

// TestStreamingE2E_MediaRecorderPushToTwoListeners exercises the browser
// contract: metadata is separated from the first recorder blob, then both late
// listeners resume from the next WebM Cluster sent through POST /push.
func TestStreamingE2E_MediaRecorderPushToTwoListeners(t *testing.T) {
	hub := streaming.NewHub()
	h := handler.NewStreamHandler(
		usecase.NewStreamUseCase(ownedLiveRepo(), nil),
		handler.WithAudioStreaming(hub, time.Minute, 5*time.Second, time.Second, 1<<20, 1024, 64),
	)
	engine := gin.New()
	engine.Use(func(c *gin.Context) {
		c.Set("claims", &entity.JWTClaims{UserID: "owner", Role: entity.RoleDiffuseur})
		c.Next()
	})
	engine.POST("/streams/:id/push", h.PushAudio)
	engine.GET("/streams/:id/audio", h.Audio)
	srv := httptest.NewServer(engine)
	defer srv.Close()
	defer hub.CloseStream("stream-1")

	push := func(payload []byte) {
		t.Helper()
		request, err := http.NewRequest(http.MethodPost, srv.URL+"/streams/stream-1/push", bytes.NewReader(payload))
		if err != nil {
			t.Fatalf("create push request: %v", err)
		}
		request.Header.Set("Content-Type", "audio/webm;codecs=opus")
		request.Header.Set("X-Stream-Session-ID", testActiveSessionID)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("push audio: %v", err)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			t.Fatalf("push status = %d, want 204", response.StatusCode)
		}
	}

	initSegment := []byte("WEBM-INIT")
	firstBlob := append(append([]byte{}, initSegment...), testWebMCluster("OLD-AUDIO")...)
	push(firstBlob)
	listeners := make([]*http.Response, 2)
	for i := range listeners {
		response, err := http.Get(srv.URL + "/streams/stream-1/audio")
		if err != nil {
			t.Fatalf("listener %d connect: %v", i+1, err)
		}
		listeners[i] = response
		defer response.Body.Close()
		gotInit := make([]byte, len(initSegment))
		if _, err := io.ReadFull(response.Body, gotInit); err != nil {
			t.Fatalf("listener %d init: %v", i+1, err)
		}
		if !bytes.Equal(gotInit, initSegment) {
			t.Fatalf("listener %d init = %q, want %q", i+1, gotInit, initSegment)
		}
	}

	mediaChunk := testWebMCluster("WEBM-MEDIA")
	push(mediaChunk)
	for i, response := range listeners {
		got := make([]byte, len(mediaChunk))
		if _, err := io.ReadFull(response.Body, got); err != nil {
			t.Fatalf("listener %d media: %v", i+1, err)
		}
		if !bytes.Equal(got, mediaChunk) {
			t.Fatalf("listener %d media = %q, want %q", i+1, got, mediaChunk)
		}
	}
}
