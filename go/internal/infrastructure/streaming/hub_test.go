package streaming_test

import (
	"errors"
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
	// InitSegment also returns a copy, so listeners cannot mutate the cache.
	returned := hub.InitSegment("stream-B")
	returned[0] = 7
	if cached := hub.InitSegment("stream-B"); cached[0] != 1 {
		t.Fatalf("cache aliased returned slice: got %v", cached)
	}
}

func TestHubRegisterWithInitAndChunkPublisher(t *testing.T) {
	hub := streaming.NewHub()
	if err := hub.OpenChunkPublisher("stream-A", "audio/webm; codecs=opus"); err != nil {
		t.Fatal(err)
	}
	// Repeated blobs reuse the same logical publisher, but changing formats or
	// opening a continuous producer at the same time is rejected.
	if err := hub.OpenChunkPublisher("stream-A", "audio/webm; codecs=opus"); err != nil {
		t.Fatalf("reuse chunk publisher: %v", err)
	}
	if err := hub.OpenChunkPublisher("stream-A", "audio/mpeg"); !errors.Is(err, streaming.ErrPublisherFormatChanged) {
		t.Fatalf("format change error = %v", err)
	}
	if _, err := hub.OpenPublisher("stream-A", "audio/webm"); !errors.Is(err, streaming.ErrPublisherActive) {
		t.Fatalf("continuous publisher error = %v", err)
	}

	initSegment := []byte("init")
	hub.SetInitSegment("stream-A", initSegment)
	client := &streaming.Client{
		ID:       "connection-1",
		UserID:   "user-1",
		StreamID: "stream-A",
		Send:     make(chan []byte, 1),
	}
	got, err := hub.RegisterWithInit(client)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(initSegment) {
		t.Fatalf("RegisterWithInit() = %q, want %q", got, initSegment)
	}

	hub.ClosePublisher("stream-A")
	if _, open := <-client.Send; open {
		t.Fatal("ClosePublisher did not end the listener response")
	}
}

func TestHubOwnedChunkPublisherRejectsLateAndForeignChunks(t *testing.T) {
	hub := streaming.NewHub()
	const contentType = "audio/webm; codecs=opus"
	if err := hub.ActivateStreamSession("stream-A", "owner", "session-1"); err != nil {
		t.Fatal(err)
	}
	if err := hub.OpenOwnedChunkPublisher("stream-A", "owner", "session-1", contentType); err != nil {
		t.Fatal(err)
	}
	if !hub.OwnsChunkPublisher("stream-A", "owner", "session-1", contentType) {
		t.Fatal("owner publisher session was not retained")
	}
	if hub.OwnsChunkPublisher("stream-A", "other", "session-1", contentType) {
		t.Fatal("foreign broadcaster matched the publisher session")
	}
	if _, _, err := hub.BroadcastOwnedChunk("stream-A", "other", "session-1", contentType, []byte("x")); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("foreign chunk error = %v, want ErrPublisherNotActive", err)
	}
	if _, _, err := hub.BroadcastOwnedChunk("stream-A", "owner", "stale-session", contentType, []byte("x")); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("stale session error = %v, want ErrPublisherNotActive", err)
	}

	hub.CloseStream("stream-A")
	if err := hub.OpenOwnedChunkPublisher("stream-A", "owner", "session-1", contentType); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("late publisher reopen error = %v, want ErrPublisherNotActive", err)
	}
	if _, _, err := hub.BroadcastOwnedChunk("stream-A", "owner", "session-1", contentType, []byte("late")); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("late chunk error = %v, want ErrPublisherNotActive", err)
	}
	if got := hub.InitSegment("stream-A"); got != nil {
		t.Fatalf("late chunk recreated init segment: %v", got)
	}

	if err := hub.ActivateStreamSession("stream-A", "owner", "session-2"); err != nil {
		t.Fatal(err)
	}
	if err := hub.OpenOwnedChunkPublisher("stream-A", "owner", "session-2", contentType); err != nil {
		t.Fatalf("new session could not publish: %v", err)
	}
	if err := hub.CloseStreamSession("stream-A", "owner", "session-1"); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("stale stop error = %v, want ErrPublisherNotActive", err)
	}
	if !hub.OwnsChunkPublisher("stream-A", "owner", "session-2", contentType) {
		t.Fatal("stale stop closed the newer publisher session")
	}
}

func TestHubCloseOwnedPublisherCannotCloseNewerSession(t *testing.T) {
	hub := streaming.NewHub()
	const contentType = "audio/webm; codecs=opus"
	if err := hub.ActivateStreamSession("stream-A", "owner", "session-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := hub.OpenOwnedPublisher("stream-A", "owner", "session-1", contentType); err != nil {
		t.Fatal(err)
	}

	if err := hub.ActivateStreamSession("stream-A", "owner", "session-2"); err != nil {
		t.Fatal(err)
	}
	if _, err := hub.OpenOwnedPublisher("stream-A", "owner", "session-2", contentType); err != nil {
		t.Fatal(err)
	}
	if hub.CloseOwnedPublisher("stream-A", "owner", "session-1") {
		t.Fatal("stale continuous publisher cleanup closed the newer session")
	}
	if _, active := hub.ContentType("stream-A"); !active {
		t.Fatal("newer continuous publisher is no longer active")
	}

	hub.CloseStream("stream-A")
}
