package streaming

import (
	"sync"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

// Client represents a connected listener on a stream.
type Client struct {
	UserID   string
	StreamID string
	Send     chan []byte
}

// Hub manages active streams and their connected listeners.
// Uses goroutines and channels, with no external dependency.
type Hub struct {
	mu              sync.RWMutex
	streams         map[string]map[string]*Client
	contentTypes    map[string]string // streamID → audio MIME type set by broadcaster
	initSegments    map[string][]byte // streamID → first chunk (WebM header) for late joiners
	userConnections map[string]int
}

// NewHub creates a new streaming Hub.
func NewHub() *Hub {
	return &Hub{
		streams:         make(map[string]map[string]*Client),
		contentTypes:    make(map[string]string),
		initSegments:    make(map[string][]byte),
		userConnections: make(map[string]int),
	}
}

// SetContentType records the audio MIME type pushed by the broadcaster.
func (h *Hub) SetContentType(streamID, contentType string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.contentTypes[streamID] = contentType
}

// ContentType returns the stored MIME type, defaulting to audio/webm; codecs=opus
// (what Chrome MediaRecorder produces — supported by all modern browsers).
func (h *Hub) ContentType(streamID string) string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if ct, ok := h.contentTypes[streamID]; ok && ct != "" {
		return ct
	}
	return "audio/webm; codecs=opus"
}

// SetInitSegment caches the first audio chunk for a stream (WebM EBML header +
// Tracks element). Listeners who join after the stream started need it so the
// browser can initialise its decoder. Only the first call per stream is kept.
func (h *Hub) SetInitSegment(streamID string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, already := h.initSegments[streamID]; !already {
		buf := make([]byte, len(data))
		copy(buf, data)
		h.initSegments[streamID] = buf
	}
}

// InitSegment returns the cached init segment for a stream, or nil.
func (h *Hub) InitSegment(streamID string) []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.initSegments[streamID]
}

// Register adds a listener to a stream.
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.streams[client.StreamID]; !ok {
		h.streams[client.StreamID] = make(map[string]*Client)
	}
	if _, exists := h.streams[client.StreamID][client.UserID]; exists {
		h.streams[client.StreamID][client.UserID] = client
		return
	}

	h.streams[client.StreamID][client.UserID] = client
	telemetry.ListenersPerStream.WithLabelValues(client.StreamID).Inc()
	if h.userConnections[client.UserID] == 0 {
		telemetry.OnlineUsers.Inc()
	}
	h.userConnections[client.UserID]++
}

// Unregister removes a listener from a stream.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	listeners, ok := h.streams[client.StreamID]
	if !ok {
		return
	}
	storedClient, exists := listeners[client.UserID]
	if !exists {
		return
	}

	delete(listeners, client.UserID)
	close(storedClient.Send)
	telemetry.ListenersPerStream.WithLabelValues(client.StreamID).Dec()
	telemetry.ListenerDisconnectTotal.Inc()

	if h.userConnections[client.UserID] > 0 {
		h.userConnections[client.UserID]--
		if h.userConnections[client.UserID] == 0 {
			delete(h.userConnections, client.UserID)
			telemetry.OnlineUsers.Dec()
		}
	}
	if len(listeners) == 0 {
		delete(h.streams, client.StreamID)
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
			// Listener too slow: drop packet without blocking the stream.
		}
	}
}

// ListenerCount returns the number of active listeners on a stream.
func (h *Hub) ListenerCount(streamID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.streams[streamID])
}
