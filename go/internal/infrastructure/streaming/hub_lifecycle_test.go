package streaming_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
)

func TestHubCloseStreamCancelsPublisherAndListeners(t *testing.T) {
	hub := streaming.NewHub()
	publisherCtx, err := hub.OpenPublisher("stream-1", "audio/mpeg")
	if err != nil {
		t.Fatalf("OpenPublisher() error = %v", err)
	}
	if _, err := hub.OpenPublisher("stream-1", "audio/mpeg"); !errors.Is(err, streaming.ErrPublisherActive) {
		t.Fatalf("second OpenPublisher() error = %v, want ErrPublisherActive", err)
	}

	client := &streaming.Client{
		ID:       "connection-1",
		UserID:   "user-1",
		StreamID: "stream-1",
		Send:     make(chan []byte, 1),
	}
	if err := hub.Register(client); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	hub.CloseStream("stream-1")

	select {
	case <-publisherCtx.Done():
	default:
		t.Fatal("publisher context was not cancelled")
	}
	if _, open := <-client.Send; open {
		t.Fatal("listener channel is still open")
	}
	if got := hub.ListenerCount("stream-1"); got != 0 {
		t.Fatalf("ListenerCount() = %d, want 0", got)
	}
}

func TestHubShutdownRejectsNewConnections(t *testing.T) {
	hub := streaming.NewHub()
	hub.Shutdown()
	hub.Shutdown() // idempotent

	client := &streaming.Client{UserID: "u", StreamID: "s", Send: make(chan []byte, 1)}
	if err := hub.Register(client); !errors.Is(err, streaming.ErrHubClosed) {
		t.Fatalf("Register() error = %v, want ErrHubClosed", err)
	}
	if _, err := hub.OpenPublisher("s", "audio/aac"); !errors.Is(err, streaming.ErrHubClosed) {
		t.Fatalf("OpenPublisher() error = %v, want ErrHubClosed", err)
	}
}

func TestHubPublisherLifecycle(t *testing.T) {
	hub := streaming.NewHub()
	ctx, err := hub.OpenPublisher("stream-1", "audio/aac")
	if err != nil {
		t.Fatal(err)
	}
	if contentType, ok := hub.ContentType("stream-1"); !ok || contentType != "audio/aac" {
		t.Fatalf("ContentType() = (%q, %v), want (audio/aac, true)", contentType, ok)
	}
	hub.ClosePublisher("stream-1")
	hub.ClosePublisher("stream-1") // idempotent
	select {
	case <-ctx.Done():
	default:
		t.Fatal("ClosePublisher did not cancel context")
	}
	if _, ok := hub.ContentType("stream-1"); ok {
		t.Fatal("publisher remains visible after close")
	}

	ctx, err = hub.OpenPublisher("stream-2", "audio/mpeg")
	if err != nil {
		t.Fatal(err)
	}
	hub.Shutdown()
	select {
	case <-ctx.Done():
	default:
		t.Fatal("Shutdown did not cancel publisher")
	}
}

func BenchmarkHubBroadcast(b *testing.B) {
	for _, listenerCount := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("listeners_%d", listenerCount), func(b *testing.B) {
			hub := streaming.NewHub()
			clients := make([]*streaming.Client, listenerCount)
			for i := range clients {
				clients[i] = &streaming.Client{
					ID:       fmt.Sprintf("connection-%d", i),
					UserID:   fmt.Sprintf("user-%d", i),
					StreamID: "benchmark",
					Send:     make(chan []byte, 1),
				}
				if err := hub.Register(clients[i]); err != nil {
					b.Fatal(err)
				}
			}
			packet := make([]byte, 32<<10)
			b.SetBytes(int64(len(packet) * listenerCount))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				delivered, dropped := hub.Broadcast("benchmark", packet)
				if delivered != listenerCount || dropped != 0 {
					b.Fatalf("Broadcast() = (%d, %d), want (%d, 0)", delivered, dropped, listenerCount)
				}
				for _, client := range clients {
					<-client.Send
				}
			}
			b.StopTimer()
			hub.Shutdown()
		})
	}
}
