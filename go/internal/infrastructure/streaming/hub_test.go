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
