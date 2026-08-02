package streaming_test

import (
	"testing"
	"time"

	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
)

func TestHub_Broadcast_DeliversToClient(t *testing.T) {
	h := streaming.NewHub()
	c := &streaming.Client{UserID: "u1", StreamID: "s1", Send: make(chan []byte, 1)}
	h.Register(c)

	h.Broadcast("s1", []byte("hello"))

	select {
	case msg := <-c.Send:
		if string(msg) != "hello" {
			t.Errorf("Broadcast delivered %q, want hello", msg)
		}
	case <-time.After(time.Second):
		t.Fatal("Broadcast did not deliver a message")
	}
}

func TestHub_Broadcast_FullChannelDoesNotBlock(t *testing.T) {
	h := streaming.NewHub()
	c := &streaming.Client{UserID: "u1", StreamID: "s1", Send: make(chan []byte, 1)}
	h.Register(c)

	// Fill the buffer so the next send would block.
	c.Send <- []byte("first")

	done := make(chan struct{})
	go func() {
		h.Broadcast("s1", []byte("dropped"))
		close(done)
	}()

	select {
	case <-done:
		// Broadcast returned without blocking — the slow-listener packet was dropped.
	case <-time.After(time.Second):
		t.Fatal("Broadcast blocked on a full channel")
	}
}

func TestHub_Broadcast_UnknownStream(t *testing.T) {
	h := streaming.NewHub()
	// Should be a no-op, no panic.
	h.Broadcast("nope", []byte("x"))
}
