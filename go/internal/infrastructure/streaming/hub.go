package streaming

import "sync"

// Client represents a connected listener on a stream.
type Client struct {
	UserID   string
	StreamID string
	Send     chan []byte
}

// Hub manages active streams and their connected listeners.
// Uses goroutines + channels — no external dependencies.
// Sprint 1 — US-003.
type Hub struct {
	mu      sync.RWMutex
	streams map[string]map[string]*Client // streamID → userID → Client
}

// NewHub creates a new streaming Hub.
func NewHub() *Hub {
	return &Hub{
		streams: make(map[string]map[string]*Client),
	}
}

// Register adds a listener to a stream.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.streams[client.StreamID]; !ok {
		h.streams[client.StreamID] = make(map[string]*Client)
	}
	h.streams[client.StreamID][client.UserID] = client
}

// Unregister removes a listener from a stream.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if listeners, ok := h.streams[client.StreamID]; ok {
		delete(listeners, client.UserID)
		close(client.Send)
	}
}

// Broadcast sends data to all listeners of a stream.
func (h *Hub) Broadcast(streamID string, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.streams[streamID] {
		select {
		case client.Send <- data:
		default:
			// Listener too slow — drop packet (non-blocking)
		}
	}
}

// ListenerCount returns the number of active listeners on a stream.
func (h *Hub) ListenerCount(streamID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.streams[streamID])
}
