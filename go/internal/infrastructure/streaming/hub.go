package streaming

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

var (
	ErrHubClosed              = errors.New("streaming: hub closed")
	ErrPublisherActive        = errors.New("streaming: publisher already active")
	ErrPublisherFormatChanged = errors.New("streaming: publisher content type changed")
	ErrPublisherNotActive     = errors.New("streaming: publisher is not active")
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
	contentType   string
	broadcasterID string
	sessionID     string
	cancel        context.CancelFunc
	startedAt     time.Time
	chunked       bool
}

type authorizedSession struct {
	broadcasterID string
	sessionID     string
	revoked       bool
}

// Hub manages active streams and their connected listeners.
// Uses goroutines and channels, with no external dependency.
type Hub struct {
	mu              sync.RWMutex
	streams         map[string]map[string]*Client
	initSegments    map[string][]byte // streamID → WebM metadata before the first Cluster
	initCandidates  map[string][]byte // bounded metadata bytes until the first Cluster is found
	userConnections map[string]int
	publishers      map[string]publisher
	sessions        map[string]authorizedSession
	closed          bool
}

// NewHub creates a new streaming Hub.
func NewHub() *Hub {
	return &Hub{
		streams:         make(map[string]map[string]*Client),
		initSegments:    make(map[string][]byte),
		initCandidates:  make(map[string][]byte),
		userConnections: make(map[string]int),
		publishers:      make(map[string]publisher),
		sessions:        make(map[string]authorizedSession),
	}
}

// ActivateStreamSession atomically replaces the session allowed to publish on
// a persistent stream. Any publisher and listeners from the preceding session
// are disconnected before the new session can accept audio.
func (h *Hub) ActivateStreamSession(streamID, broadcasterID, sessionID string) error {
	if streamID == "" || broadcasterID == "" || sessionID == "" {
		return ErrPublisherNotActive
	}
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return ErrHubClosed
	}
	pub, hasPublisher := h.closeStreamLocked(streamID)
	h.sessions[streamID] = authorizedSession{
		broadcasterID: broadcasterID,
		sessionID:     sessionID,
	}
	h.mu.Unlock()
	h.finishPublisher(streamID, pub, hasPublisher)
	return nil
}

// AuthorizeStreamSession initializes the in-memory session after a process
// restart, once the database has confirmed it. A revoked or superseded session
// can never claim the stream again; only ActivateStreamSession may replace it.
func (h *Hub) AuthorizeStreamSession(streamID, broadcasterID, sessionID string) error {
	if streamID == "" || broadcasterID == "" || sessionID == "" {
		return ErrPublisherNotActive
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return ErrHubClosed
	}
	current, exists := h.sessions[streamID]
	if !exists {
		h.sessions[streamID] = authorizedSession{
			broadcasterID: broadcasterID,
			sessionID:     sessionID,
		}
		return nil
	}
	if current.revoked || current.broadcasterID != broadcasterID || current.sessionID != sessionID {
		return ErrPublisherNotActive
	}
	return nil
}

func (h *Hub) sessionAuthorizedLocked(streamID, broadcasterID, sessionID string) bool {
	current, exists := h.sessions[streamID]
	return exists && !current.revoked &&
		current.broadcasterID == broadcasterID && current.sessionID == sessionID
}

// Register adds a listener to a stream. A duplicate connection for the same key
// (client ID, or user ID when no ID is set) replaces the previous one.
func (h *Hub) Register(client *Client) error {
	_, err := h.RegisterWithInit(client)
	return err
}

// RegisterWithInit atomically registers a listener and snapshots the cached
// metadata segment. The shared lock defines whether the listener receives the
// stream from its beginning or bootstraps later at a new Cluster boundary.
func (h *Hub) RegisterWithInit(client *Client) ([]byte, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return nil, ErrHubClosed
	}
	if client.Send == nil {
		return nil, errors.New("streaming: client send channel is nil")
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
		return append([]byte(nil), h.initSegments[client.StreamID]...), nil
	}

	h.streams[client.StreamID][key] = client
	telemetry.ListenersPerStream.WithLabelValues(client.StreamID).Inc()
	if h.userConnections[client.UserID] == 0 {
		telemetry.OnlineUsers.Inc()
	}
	h.userConnections[client.UserID]++
	return append([]byte(nil), h.initSegments[client.StreamID]...), nil
}

// SetInitSegment explicitly stores decoder metadata. Browser ingestion uses
// cacheWebMInitializationLocked instead so media from the first recorder blob
// is never mistaken for initialization data.
func (h *Hub) SetInitSegment(streamID string, data []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, already := h.initSegments[streamID]; !already {
		buf := make([]byte, len(data))
		copy(buf, data)
		h.initSegments[streamID] = buf
		delete(h.initCandidates, streamID)
	}
}

const maxWebMInitializationBytes = 1 << 20

func (h *Hub) cacheWebMInitializationLocked(streamID string, data []byte) {
	if _, ready := h.initSegments[streamID]; ready {
		return
	}
	candidate := make([]byte, 0, len(h.initCandidates[streamID])+len(data))
	candidate = append(candidate, h.initCandidates[streamID]...)
	candidate = append(candidate, data...)
	if initialization, found := webMInitializationPrefix(candidate); found {
		h.initSegments[streamID] = initialization
		delete(h.initCandidates, streamID)
		return
	}
	if len(candidate) <= maxWebMInitializationBytes {
		h.initCandidates[streamID] = candidate
	} else {
		delete(h.initCandidates, streamID)
	}
}

// InitSegment returns the cached init segment for a stream, or nil.
func (h *Hub) InitSegment(streamID string) []byte {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return append([]byte(nil), h.initSegments[streamID]...)
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

// Broadcast sends data to all listeners of a stream.
func (h *Hub) Broadcast(streamID string, data []byte) (delivered, dropped int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.broadcastLocked(streamID, data)
}

// BroadcastWithInit atomically extracts WebM decoder metadata and fans out the
// original continuous bytes. The cached value never includes Cluster media.
func (h *Hub) BroadcastWithInit(streamID string, data []byte) (delivered, dropped int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cacheWebMInitializationLocked(streamID, data)
	return h.broadcastLocked(streamID, data)
}

// broadcastLocked requires the Hub write lock. Losing arbitrary bytes corrupts
// a dependent WebM stream, so a slow listener is disconnected instead of
// receiving later data after a dropped packet.
func (h *Hub) broadcastLocked(streamID string, data []byte) (delivered, dropped int) {
	// One immutable copy detaches the packet from the broadcaster's reusable
	// read buffer. All listener queues may safely share that same byte slice.
	packet := append([]byte(nil), data...)
	listeners := h.streams[streamID]
	for key, client := range listeners {
		select {
		case client.Send <- packet:
			delivered++
		default:
			dropped++
			delete(listeners, key)
			h.disconnectLocked(client)
		}
	}
	if len(listeners) == 0 {
		delete(h.streams, streamID)
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
	delete(h.initSegments, streamID)
	delete(h.initCandidates, streamID)
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(1)
	return ctx, nil
}

// OpenOwnedPublisher reserves the continuous ingestion slot for an authorized
// broadcast session. The session gate prevents a request that raced with Stop
// from reopening a publisher after CloseStream.
func (h *Hub) OpenOwnedPublisher(
	streamID, broadcasterID, sessionID, contentType string,
) (context.Context, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return nil, ErrHubClosed
	}
	if !h.sessionAuthorizedLocked(streamID, broadcasterID, sessionID) {
		return nil, ErrPublisherNotActive
	}
	if _, exists := h.publishers[streamID]; exists {
		return nil, ErrPublisherActive
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.publishers[streamID] = publisher{
		contentType:   contentType,
		broadcasterID: broadcasterID,
		sessionID:     sessionID,
		cancel:        cancel,
		startedAt:     time.Now(),
	}
	delete(h.initSegments, streamID)
	delete(h.initCandidates, streamID)
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(1)
	return ctx, nil
}

// OpenChunkPublisher opens, or reuses, the publisher session used by browser
// MediaRecorder uploads. Unlike the continuous PUT publisher, the session
// spans multiple short HTTP requests and is closed by Stop/CloseStream.
func (h *Hub) OpenChunkPublisher(streamID, contentType string) error {
	return h.OpenOwnedChunkPublisher(streamID, "", "", contentType)
}

// OpenOwnedChunkPublisher opens, or reuses, a browser publisher owned by the
// authenticated broadcaster. Ownership lets subsequent short chunk requests
// stay entirely on the in-memory data plane after the first database check.
func (h *Hub) OpenOwnedChunkPublisher(streamID, broadcasterID, sessionID, contentType string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return ErrHubClosed
	}
	// The blank pair is retained for the low-level compatibility wrapper used
	// by Hub tests and non-authenticated internal callers.
	if (broadcasterID != "" || sessionID != "") &&
		!h.sessionAuthorizedLocked(streamID, broadcasterID, sessionID) {
		return ErrPublisherNotActive
	}
	if current, exists := h.publishers[streamID]; exists {
		if !current.chunked {
			return ErrPublisherActive
		}
		if current.contentType != contentType {
			return ErrPublisherFormatChanged
		}
		if current.broadcasterID != broadcasterID {
			return ErrPublisherActive
		}
		if current.sessionID != sessionID {
			return ErrPublisherActive
		}
		return nil
	}
	_, cancel := context.WithCancel(context.Background())
	h.publishers[streamID] = publisher{
		contentType:   contentType,
		broadcasterID: broadcasterID,
		sessionID:     sessionID,
		cancel:        cancel,
		startedAt:     time.Now(),
		chunked:       true,
	}
	delete(h.initSegments, streamID)
	delete(h.initCandidates, streamID)
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(1)
	return nil
}

// OwnsChunkPublisher reports whether this exact authenticated publisher
// session is already open. It is used to avoid a database lookup per blob.
func (h *Hub) OwnsChunkPublisher(streamID, broadcasterID, sessionID, contentType string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	current, exists := h.publishers[streamID]
	return exists && current.chunked &&
		current.broadcasterID == broadcasterID &&
		current.sessionID == sessionID &&
		current.contentType == contentType
}

// BroadcastOwnedChunk verifies that the publisher still exists while holding
// the same lock used by CloseStream, then atomically caches and broadcasts the
// payload. A request that finishes after Stop therefore cannot revive audio.
func (h *Hub) BroadcastOwnedChunk(
	streamID, broadcasterID, sessionID, contentType string,
	data []byte,
) (delivered, dropped int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.closed {
		return 0, 0, ErrHubClosed
	}
	current, exists := h.publishers[streamID]
	if !exists || !current.chunked || current.broadcasterID != broadcasterID || current.sessionID != sessionID {
		return 0, 0, ErrPublisherNotActive
	}
	if current.contentType != contentType {
		return 0, 0, ErrPublisherFormatChanged
	}
	h.cacheWebMInitializationLocked(streamID, data)
	delivered, dropped = h.broadcastLocked(streamID, data)
	return delivered, dropped, nil
}

// ClosePublisher releases a broadcaster slot. It is idempotent.
func (h *Hub) ClosePublisher(streamID string) {
	h.mu.Lock()
	pub, exists := h.publishers[streamID]
	if exists {
		delete(h.publishers, streamID)
	}
	listeners := h.streams[streamID]
	delete(h.streams, streamID)
	delete(h.initSegments, streamID)
	delete(h.initCandidates, streamID)
	for _, client := range listeners {
		h.disconnectLocked(client)
	}
	h.mu.Unlock()
	if !exists {
		return
	}
	pub.cancel()
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(0)
	telemetry.BroadcasterSessionDuration.Observe(time.Since(pub.startedAt).Seconds())
}

// CloseOwnedPublisher closes only the continuous publisher belonging to the
// named broadcast session. A deferred cleanup from an older request must not
// tear down a newer session which reused the stable stream ID.
func (h *Hub) CloseOwnedPublisher(streamID, broadcasterID, sessionID string) bool {
	h.mu.Lock()
	pub, exists := h.publishers[streamID]
	if !exists || pub.chunked || pub.broadcasterID != broadcasterID || pub.sessionID != sessionID {
		h.mu.Unlock()
		return false
	}
	delete(h.publishers, streamID)
	listeners := h.streams[streamID]
	delete(h.streams, streamID)
	delete(h.initSegments, streamID)
	delete(h.initCandidates, streamID)
	for _, client := range listeners {
		h.disconnectLocked(client)
	}
	h.mu.Unlock()
	h.finishPublisher(streamID, pub, true)
	return true
}

// ContentType returns the active publisher's media type and whether a publisher
// is currently connected.
func (h *Hub) ContentType(streamID string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	pub, ok := h.publishers[streamID]
	return pub.contentType, ok
}

// CloseStream cancels ingestion, disconnects every listener and clears the
// cached WebM metadata when a broadcaster stops. Closing the channels also lets
// the HTTP handlers finish their chunked responses so listeners receive a real
// end-of-stream.
func (h *Hub) CloseStream(streamID string) {
	h.mu.Lock()
	pub, hasPublisher := h.closeStreamLocked(streamID)
	h.sessions[streamID] = authorizedSession{revoked: true}
	h.mu.Unlock()
	h.finishPublisher(streamID, pub, hasPublisher)
}

// CloseStreamSession revokes and closes only the named session. This keeps a
// delayed Stop response from tearing down a newer session that reused the same
// persistent stream ID.
func (h *Hub) CloseStreamSession(streamID, broadcasterID, sessionID string) error {
	if streamID == "" || broadcasterID == "" || sessionID == "" {
		return ErrPublisherNotActive
	}
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return ErrHubClosed
	}
	current, exists := h.sessions[streamID]
	if exists && (current.broadcasterID != broadcasterID || current.sessionID != sessionID) {
		h.mu.Unlock()
		return ErrPublisherNotActive
	}
	pub, hasPublisher := h.closeStreamLocked(streamID)
	h.sessions[streamID] = authorizedSession{
		broadcasterID: broadcasterID,
		sessionID:     sessionID,
		revoked:       true,
	}
	h.mu.Unlock()
	h.finishPublisher(streamID, pub, hasPublisher)
	return nil
}

// closeStreamLocked detaches all volatile media state. The caller owns h.mu.
func (h *Hub) closeStreamLocked(streamID string) (publisher, bool) {
	pub, hasPublisher := h.publishers[streamID]
	if hasPublisher {
		delete(h.publishers, streamID)
	}
	listeners := h.streams[streamID]
	delete(h.streams, streamID)
	delete(h.initSegments, streamID)
	delete(h.initCandidates, streamID)
	for _, client := range listeners {
		h.disconnectLocked(client)
	}
	return pub, hasPublisher
}

func (h *Hub) finishPublisher(streamID string, pub publisher, exists bool) {
	if !exists {
		return
	}
	pub.cancel()
	telemetry.AudioBroadcasters.WithLabelValues(streamID).Set(0)
	telemetry.BroadcasterSessionDuration.Observe(time.Since(pub.startedAt).Seconds())
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
