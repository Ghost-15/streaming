package streaming

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

var (
	ErrHubClosed       = errors.New("streaming: hub closed")
	ErrPublisherActive = errors.New("streaming: publisher already active")
)

// Client represents a connected listener on a stream.
type Client struct {
	ID       string
	UserID   string
	StreamID string
	Send     chan []byte
	joinedAt time.Time
}

type publisher struct {
	contentType string
	cancel      context.CancelFunc
	startedAt   time.Time
}

// Hub manages active streams and their connected listeners.
// Uses goroutines and channels, with no external dependency.
type Hub struct {
	mu              sync.RWMutex
	streams         map[string]map[string]*Client
	contentTypes    map[string]string // streamID → audio MIME type set by broadcaster
	initSegments    map[string][]byte // streamID → first chunk (WebM header) for late joiners
	userConnections map[string]int
	publishers      map[string]publisher
	closed          bool
}

// NewHub creates a new streaming Hub.
func NewHub() *Hub {
	return &Hub{
		streams:         make(map[string]map[string]*Client),
		contentTypes:    make(map[string]string),
		initSegments:    make(map[string][]byte),
		userConnections: make(map[string]int),
		publishers:      make(map[string]publisher),
	}
}

// Register adds a listener to a stream.
func (h *Hub) Register(client *Client) error {
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

// Register adds a listener to a stream and atomically snapshots its cached
// initialization segment. This prevents a first chunk from being both returned
// here and queued concurrently by Broadcast.
func (h *Hub) Register(client *Client) []byte {
	h.mu.Lock()
	defer h.mu.Unlock()
	initSegment := append([]byte(nil), h.initSegments[client.StreamID]...)

	if h.closed {
		return ErrHubClosed
	}
	if client.Send == nil {
		return errors.New("streaming: client send channel is nil")
	}
	if client.joinedAt.IsZero() {
		client.joinedAt = time.Now()
	}
	key := client.ID
	if key == "" {
		// Backward-compatible key for callers that do not need multiple
		// simultaneous devices for one user.
		key = client.UserID
	}
	if _, ok := h.streams[client.StreamID]; !ok {
		h.streams[client.StreamID] = make(map[string]*Client)
	}
	if previous, exists := h.streams[client.StreamID][key]; exists {
		close(previous.Send)
		h.streams[client.StreamID][key] = client
		return nil
	}

	h.streams[client.StreamID][key] = client
	telemetry.ListenersPerStream.WithLabelValues(client.StreamID).Inc()
	if h.userConnections[client.UserID] == 0 {
		telemetry.OnlineUsers.Inc()
	}
	h.userConnections[client.UserID]++
	return nil
}

// Unregister removes a listener from a stream.
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	listeners, ok := h.streams[client.StreamID]
	if !ok {
		return
	}
	key := client.ID
	if key == "" {
		key = client.UserID
	}
	storedClient, exists := listeners[key]
	if !exists {
		return
	}

	delete(listeners, key)
	h.disconnectLocked(storedClient)
	if len(listeners) == 0 {
		delete(h.streams, client.StreamID)
	}
}

// CloseStream disconnects every listener and removes the cached WebM metadata
// when a broadcaster stops. Closing the channels also lets the HTTP handlers
// finish their chunked responses, so listeners receive a real end-of-stream.
func (h *Hub) CloseStream(streamID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	listeners := h.streams[streamID]
	delete(h.streams, streamID)
	delete(h.contentTypes, streamID)
	delete(h.initSegments, streamID)

	for _, client := range listeners {
		close(client.Send)
		telemetry.ListenersPerStream.WithLabelValues(streamID).Dec()
		telemetry.ListenerDisconnectTotal.Inc()

		if h.userConnections[client.UserID] > 0 {
			h.userConnections[client.UserID]--
			if h.userConnections[client.UserID] == 0 {
				delete(h.userConnections, client.UserID)
				telemetry.OnlineUsers.Dec()
			}
		}
	}
}

// Broadcast sends data to all listeners of a stream.
func (h *Hub) Broadcast(streamID string, data []byte) (delivered, dropped int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	// One immutable copy detaches the packet from the broadcaster's reusable
	// read buffer. All listener queues may safely share that same byte slice.
	packet := append([]byte(nil), data...)
	for _, client := range h.streams[streamID] {
		select {
		case client.Send <- packet:
			delivered++
		default:
			dropped++
		}
	}
	telemetry.AudioIngestBytesTotal.WithLabelValues(streamID).Add(float64(len(data)))
	telemetry.AudioChunksTotal.WithLabelValues(streamID, "ingest").Inc()
	telemetry.AudioChunkSizeBytes.Observe(float64(len(data)))
	if dropped > 0 {
		telemetry.AudioDroppedChunksTotal.WithLabelValues(streamID).Add(float64(dropped))
	}
	return delivered, dropped
}

// ListenerCount returns the number of active listeners on a stream.
func (h *Hub) ListenerCount(streamID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.streams[streamID])
}

// OpenPublisher reserves a stream for one broadcaster and returns a context
// cancelled by CloseStream or Shutdown.
func (h *Hub) OpenPublisher(streamID, contentType string) (context.Context, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil, ErrHubClosed
	}
	if _, exists := h.publishers[streamID]; exists {
		return nil, ErrPublisherActive
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.publishers[streamID] = publisher{
		contentType: contentType,
		cancel:      cancel,
		startedAt:   time.Now(),
	}
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(1)
	return ctx, nil
}

// ClosePublisher releases a broadcaster slot. It is idempotent.
func (h *Hub) ClosePublisher(streamID string) {
	h.mu.Lock()
	pub, exists := h.publishers[streamID]
	if exists {
		delete(h.publishers, streamID)
	}
	h.mu.Unlock()
	if !exists {
		return
	}
	pub.cancel()
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(0)
	telemetry.BroadcasterSessionDuration.Observe(time.Since(pub.startedAt).Seconds())
}

// ContentType returns the active publisher's media type.
func (h *Hub) ContentType(streamID string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	pub, ok := h.publishers[streamID]
	return pub.contentType, ok
}

// CloseStream cancels ingestion and disconnects all real audio listeners.
func (h *Hub) CloseStream(streamID string) {
	h.mu.Lock()
	pub, hasPublisher := h.publishers[streamID]
	if hasPublisher {
		delete(h.publishers, streamID)
	}
	listeners := h.streams[streamID]
	delete(h.streams, streamID)
	for _, client := range listeners {
		h.disconnectLocked(client)
	}
	h.mu.Unlock()

	if hasPublisher {
		pub.cancel()
		telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(0)
		telemetry.BroadcasterSessionDuration.Observe(time.Since(pub.startedAt).Seconds())
	}
}

// Shutdown releases every long-lived connection and rejects new ones.
func (h *Hub) Shutdown() {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	h.closed = true
	publishers := h.publishers
	h.publishers = make(map[string]publisher)
	for streamID, listeners := range h.streams {
		for _, client := range listeners {
			h.disconnectLocked(client)
		}
		delete(h.streams, streamID)
	}
	h.mu.Unlock()

	for streamID, pub := range publishers {
		pub.cancel()
		telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(0)
		telemetry.BroadcasterSessionDuration.Observe(time.Since(pub.startedAt).Seconds())
	}
}

func (h *Hub) disconnectLocked(client *Client) {
	close(client.Send)
	telemetry.ListenersPerStream.WithLabelValues(client.StreamID).Dec()
	telemetry.ListenerDisconnectTotal.Inc()
	if !client.joinedAt.IsZero() {
		telemetry.ListenerSessionDuration.Observe(time.Since(client.joinedAt).Seconds())
	}
	if h.userConnections[client.UserID] > 0 {
		h.userConnections[client.UserID]--
		if h.userConnections[client.UserID] == 0 {
			delete(h.userConnections, client.UserID)
			telemetry.OnlineUsers.Dec()
		}
	}
}
