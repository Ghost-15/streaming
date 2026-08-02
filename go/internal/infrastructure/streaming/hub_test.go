package streaming_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

func TestHubMetricsTrackOnlineUsersListenersAndDisconnects(t *testing.T) {
	telemetry.ListenersPerStream.Reset()
	telemetry.OnlineUsers.Set(0)
	beforeDisconnects := testutil.ToFloat64(telemetry.ListenerDisconnectTotal)

	hub := streaming.NewHub()
	firstConnection := &streaming.Client{UserID: "user-1", StreamID: "stream-A", Send: make(chan []byte, 1)}
	duplicateConnection := &streaming.Client{UserID: "user-1", StreamID: "stream-A", Send: make(chan []byte, 1)}
	secondStreamConnection := &streaming.Client{UserID: "user-1", StreamID: "stream-B", Send: make(chan []byte, 1)}

	hub.Register(firstConnection)
	hub.Register(duplicateConnection)
	hub.Register(secondStreamConnection)

	if got := hub.ListenerCount("stream-A"); got != 1 {
		t.Fatalf("ListenerCount(stream-A) = %d, want 1", got)
	}
	if got := testutil.ToFloat64(telemetry.ListenersPerStream.WithLabelValues("stream-A")); got != 1 {
		t.Fatalf("ListenersPerStream(stream-A) = %v, want 1", got)
	}
	if got := testutil.ToFloat64(telemetry.ListenersPerStream.WithLabelValues("stream-B")); got != 1 {
		t.Fatalf("ListenersPerStream(stream-B) = %v, want 1", got)
	}
	if got := testutil.ToFloat64(telemetry.OnlineUsers); got != 1 {
		t.Fatalf("OnlineUsers = %v, want 1", got)
	}

	hub.Unregister(duplicateConnection)

	if got := testutil.ToFloat64(telemetry.ListenersPerStream.WithLabelValues("stream-A")); got != 0 {
		t.Fatalf("ListenersPerStream(stream-A) after unregister = %v, want 0", got)
	}
	if got := testutil.ToFloat64(telemetry.OnlineUsers); got != 1 {
		t.Fatalf("OnlineUsers after first unregister = %v, want 1", got)
	}

	hub.Unregister(secondStreamConnection)

	if got := testutil.ToFloat64(telemetry.OnlineUsers); got != 0 {
		t.Fatalf("OnlineUsers after final unregister = %v, want 0", got)
	}
	afterDisconnects := testutil.ToFloat64(telemetry.ListenerDisconnectTotal)
	if afterDisconnects-beforeDisconnects != 2 {
		t.Fatalf("ListenerDisconnectTotal delta = %v, want 2", afterDisconnects-beforeDisconnects)
	}
}

func TestHubCloseStreamDisconnectsListenersAndClearsMetadata(t *testing.T) {
	telemetry.ListenersPerStream.Reset()
	telemetry.OnlineUsers.Set(0)

	hub := streaming.NewHub()
	first := &streaming.Client{UserID: "user-1", StreamID: "stream-A", Send: make(chan []byte, 1)}
	second := &streaming.Client{UserID: "user-2", StreamID: "stream-A", Send: make(chan []byte, 1)}
	hub.Register(first)
	hub.Register(second)
	hub.SetInitSegment("stream-A", []byte{0x1a, 0x45, 0xdf, 0xa3})

	hub.CloseStream("stream-A")

	if got := hub.ListenerCount("stream-A"); got != 0 {
		t.Fatalf("ListenerCount(stream-A) = %d, want 0", got)
	}
	if _, open := <-first.Send; open {
		t.Fatal("first listener channel is still open")
	}
	if _, open := <-second.Send; open {
		t.Fatal("second listener channel is still open")
	}
	if got := hub.InitSegment("stream-A"); got != nil {
		t.Fatalf("InitSegment(stream-A) = %v, want nil", got)
	}
	if got := testutil.ToFloat64(telemetry.OnlineUsers); got != 0 {
		t.Fatalf("OnlineUsers after CloseStream = %v, want 0", got)
	}
}

func TestHubInitSegmentCaching(t *testing.T) {
	hub := streaming.NewHub()
	want := []byte{0x1a, 0x45, 0xdf, 0xa3}

	// Only the first cached chunk per stream is kept.
	hub.SetInitSegment("stream-A", want)
	hub.SetInitSegment("stream-A", []byte{0xff, 0xff})

	got := hub.InitSegment("stream-A")
	if len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("InitSegment = %v, want %v", got, want)
	}

	// SetInitSegment stores a copy: mutating the source must not corrupt the cache.
	src := []byte{1, 2, 3}
	hub.SetInitSegment("stream-B", src)
	src[0] = 9
	if cached := hub.InitSegment("stream-B"); cached[0] != 1 {
		t.Fatalf("cache aliased caller slice: got %v", cached)
	}
}
