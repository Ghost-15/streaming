package streaming_test

import (
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
)

func TestHubAuthorizeStreamSession(t *testing.T) {
	hub := streaming.NewHub()

	if err := hub.AuthorizeStreamSession("", "owner", "sess-1"); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("empty stream: err = %v, want ErrPublisherNotActive", err)
	}

	// First authorization records the session.
	if err := hub.AuthorizeStreamSession("stream-A", "owner", "sess-1"); err != nil {
		t.Fatalf("first authorize: %v", err)
	}
	// Re-authorizing the same identity is idempotent (process restart replay).
	if err := hub.AuthorizeStreamSession("stream-A", "owner", "sess-1"); err != nil {
		t.Fatalf("replay authorize: %v", err)
	}
	// A foreign broadcaster or a different session can never claim it.
	if err := hub.AuthorizeStreamSession("stream-A", "intruder", "sess-1"); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("foreign broadcaster: err = %v, want ErrPublisherNotActive", err)
	}
	if err := hub.AuthorizeStreamSession("stream-A", "owner", "sess-2"); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("different session: err = %v, want ErrPublisherNotActive", err)
	}

	// CloseStream revokes the session; it cannot be reauthorized afterwards.
	hub.CloseStream("stream-A")
	if err := hub.AuthorizeStreamSession("stream-A", "owner", "sess-1"); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("revoked authorize: err = %v, want ErrPublisherNotActive", err)
	}

	hub.Shutdown()
	if err := hub.AuthorizeStreamSession("stream-B", "owner", "sess-1"); !errors.Is(err, streaming.ErrHubClosed) {
		t.Fatalf("closed hub: err = %v, want ErrHubClosed", err)
	}
}

func TestHubOpenAndCloseOwnedContinuousPublisher(t *testing.T) {
	hub := streaming.NewHub()
	const (
		sid = "stream-A"
		bc  = "owner"
		ss  = "sess-1"
		ct  = "audio/mpeg"
	)
	if err := hub.ActivateStreamSession(sid, bc, ss); err != nil {
		t.Fatal(err)
	}

	// An unauthorized session cannot open the continuous publisher.
	if _, err := hub.OpenOwnedPublisher(sid, "intruder", ss, ct); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("unauthorized open: err = %v, want ErrPublisherNotActive", err)
	}
	ctx, err := hub.OpenOwnedPublisher(sid, bc, ss, ct)
	if err != nil {
		t.Fatalf("authorized open: %v", err)
	}
	if _, err := hub.OpenOwnedPublisher(sid, bc, ss, ct); !errors.Is(err, streaming.ErrPublisherActive) {
		t.Fatalf("second open: err = %v, want ErrPublisherActive", err)
	}
	if got, ok := hub.ContentType(sid); !ok || got != ct {
		t.Fatalf("ContentType = (%q, %v), want (%q, true)", got, ok, ct)
	}

	// A close from a different session must not tear down this publisher.
	if hub.CloseOwnedPublisher(sid, bc, "stale-session") {
		t.Fatal("stale session should not close the publisher")
	}
	// The owning session closes it and cancels the ingestion context.
	if !hub.CloseOwnedPublisher(sid, bc, ss) {
		t.Fatal("owning session should close the publisher")
	}
	select {
	case <-ctx.Done():
	default:
		t.Fatal("CloseOwnedPublisher did not cancel the ingestion context")
	}
	// Idempotent: nothing left to close.
	if hub.CloseOwnedPublisher(sid, bc, ss) {
		t.Fatal("second close should report false")
	}
}

func TestHubCloseOwnedPublisherIgnoresChunkPublisher(t *testing.T) {
	hub := streaming.NewHub()
	const (
		sid = "stream-A"
		bc  = "owner"
		ss  = "sess-1"
		ct  = "audio/webm; codecs=opus"
	)
	if err := hub.ActivateStreamSession(sid, bc, ss); err != nil {
		t.Fatal(err)
	}
	if err := hub.OpenOwnedChunkPublisher(sid, bc, ss, ct); err != nil {
		t.Fatal(err)
	}
	// CloseOwnedPublisher only targets the continuous publisher, never the
	// browser chunk publisher.
	if hub.CloseOwnedPublisher(sid, bc, ss) {
		t.Fatal("CloseOwnedPublisher must not close a chunk publisher")
	}
}

func TestHubCloseStreamSession(t *testing.T) {
	hub := streaming.NewHub()
	const (
		sid = "stream-A"
		bc  = "owner"
		ss  = "sess-1"
	)
	if err := hub.CloseStreamSession("", bc, ss); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("empty args: err = %v, want ErrPublisherNotActive", err)
	}

	if err := hub.ActivateStreamSession(sid, bc, ss); err != nil {
		t.Fatal(err)
	}
	// A delayed Stop from another session must not tear down this one.
	if err := hub.CloseStreamSession(sid, bc, "other-session"); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("mismatched session: err = %v, want ErrPublisherNotActive", err)
	}
	// The owning Stop revokes the session.
	if err := hub.CloseStreamSession(sid, bc, ss); err != nil {
		t.Fatalf("owning close: %v", err)
	}
	if err := hub.AuthorizeStreamSession(sid, bc, ss); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("revoked after close: err = %v, want ErrPublisherNotActive", err)
	}

	hub.Shutdown()
	if err := hub.CloseStreamSession("stream-B", bc, ss); !errors.Is(err, streaming.ErrHubClosed) {
		t.Fatalf("closed hub: err = %v, want ErrHubClosed", err)
	}
}

func TestHubBroadcastOwnedChunkErrors(t *testing.T) {
	hub := streaming.NewHub()
	const (
		sid = "stream-A"
		bc  = "owner"
		ss  = "sess-1"
		ct  = "audio/webm; codecs=opus"
	)
	if err := hub.ActivateStreamSession(sid, bc, ss); err != nil {
		t.Fatal(err)
	}
	if err := hub.OpenOwnedChunkPublisher(sid, bc, ss, ct); err != nil {
		t.Fatal(err)
	}

	if _, _, err := hub.BroadcastOwnedChunk(sid, bc, "other", ct, []byte("x")); !errors.Is(err, streaming.ErrPublisherNotActive) {
		t.Fatalf("foreign session: err = %v, want ErrPublisherNotActive", err)
	}
	if _, _, err := hub.BroadcastOwnedChunk(sid, bc, ss, "audio/mpeg", []byte("x")); !errors.Is(err, streaming.ErrPublisherFormatChanged) {
		t.Fatalf("format change: err = %v, want ErrPublisherFormatChanged", err)
	}
	// A valid chunk is accepted even with no listener connected.
	if _, _, err := hub.BroadcastOwnedChunk(sid, bc, ss, ct, []byte("data")); err != nil {
		t.Fatalf("valid chunk: %v", err)
	}

	hub.Shutdown()
	if _, _, err := hub.BroadcastOwnedChunk(sid, bc, ss, ct, []byte("x")); !errors.Is(err, streaming.ErrHubClosed) {
		t.Fatalf("closed hub: err = %v, want ErrHubClosed", err)
	}
}
