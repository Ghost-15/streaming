package handler

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// StreamHandler handles HTTP requests for live streams.
type StreamHandler struct {
	useCase       usecase.StreamUseCase
	hub           *streaming.Hub
	maxDuration   time.Duration
	idleTimeout   time.Duration
	writeTimeout  time.Duration
	leaveTimeout  time.Duration
	maxIngestSize int64
	chunkSize     int
	clientBuffer  int
	controlMu     sync.Mutex
}

type StreamHandlerOption func(*StreamHandler)

// WithAudioStreaming enables the real HTTP audio data plane. Keeping this an
// option preserves small unit-test handlers that only exercise the JSON API.
func WithAudioStreaming(
	hub *streaming.Hub,
	maxDuration, idleTimeout, writeTimeout time.Duration,
	maxIngestSize int64,
	chunkSize, clientBuffer int,
) StreamHandlerOption {
	return func(h *StreamHandler) {
		h.hub = hub
		h.maxDuration = maxDuration
		h.idleTimeout = idleTimeout
		h.writeTimeout = writeTimeout
		h.maxIngestSize = maxIngestSize
		h.chunkSize = chunkSize
		h.clientBuffer = clientBuffer
	}
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(uc usecase.StreamUseCase, options ...StreamHandlerOption) *StreamHandler {
	h := &StreamHandler{
		useCase:       uc,
		maxDuration:   6 * time.Hour,
		idleTimeout:   30 * time.Second,
		writeTimeout:  10 * time.Second,
		leaveTimeout:  3 * time.Second,
		maxIngestSize: 8 << 30,
		chunkSize:     32 << 10,
		clientBuffer:  64,
	}
	for _, option := range options {
		option(h)
	}
	return h
}

// StartRequest is the JSON body for POST /streams.
type StartRequest struct {
	Title string `json:"title" binding:"required,min=3,max=100"`
}

const maxBrowserAudioChunkSize int64 = 4 << 20

func parseAudioContentType(value string) (string, error) {
	mediaType, params, err := mime.ParseMediaType(value)
	if err != nil || (!strings.HasPrefix(mediaType, "audio/") && mediaType != "application/octet-stream") {
		return "", errors.New("Content-Type must be audio/* or application/octet-stream")
	}
	return mime.FormatMediaType(mediaType, params), nil
}

func mapStreamError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrStreamInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, usecase.ErrStreamNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
	case errors.Is(err, usecase.ErrStreamForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, usecase.ErrStreamSessionExpired):
		c.JSON(http.StatusConflict, gin.H{"error": "broadcast session expired"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func mapStreamSessionError(c *gin.Context, err error) {
	if errors.Is(err, streaming.ErrHubClosed) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
		return
	}
	c.JSON(http.StatusConflict, gin.H{"error": "broadcast session expired"})
}

// ListOwned returns every reusable live created by the authenticated
// broadcaster, including offline ones.
func (h *StreamHandler) ListOwned(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	streams, err := h.useCase.ListOwned(c.Request.Context(), claims.UserID)
	if err != nil {
		mapStreamError(c, err)
		return
	}
	c.JSON(http.StatusOK, streams)
}

// ListActive lists all currently live streams. This route is intentionally
// public so anonymous users can browse public live streams.
func (h *StreamHandler) ListActive(c *gin.Context) {
	streams, err := h.useCase.ListActive(c.Request.Context())
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("list active streams failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Int("count", len(streams)).Msg("listed active streams")
	c.JSON(http.StatusOK, streams)
}

// Start creates a new live stream for the authenticated broadcaster.
func (h *StreamHandler) Start(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	var req StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid start stream payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stream, err := h.useCase.Start(c.Request.Context(), claims.UserID, req.Title)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("start stream failed")
		mapStreamError(c, err)
		return
	}
	if h.hub != nil && stream.ActiveSessionID != nil {
		if err := h.hub.ActivateStreamSession(stream.ID, claims.UserID, *stream.ActiveSessionID); err != nil {
			_ = h.useCase.End(c.Request.Context(), stream.ID, claims.UserID, *stream.ActiveSessionID)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
			return
		}
	}
	middleware.Logger(c).Info().Str("stream_id", stream.ID).Msg("stream started")
	c.JSON(http.StatusCreated, stream)
}

// Restart begins a new broadcast session on an existing persistent stream.
func (h *StreamHandler) Restart(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	h.controlMu.Lock()
	defer h.controlMu.Unlock()
	streamID := c.Param("id")
	stream, err := h.useCase.Restart(c.Request.Context(), streamID, claims.UserID)
	if err != nil {
		mapStreamError(c, err)
		return
	}
	// Replace the previous in-memory publisher only after ownership has been
	// checked and the database has stored the new session.
	if h.hub != nil && stream.ActiveSessionID != nil {
		if err := h.hub.ActivateStreamSession(streamID, claims.UserID, *stream.ActiveSessionID); err != nil {
			_ = h.useCase.End(c.Request.Context(), streamID, claims.UserID, *stream.ActiveSessionID)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
			return
		}
	}
	c.JSON(http.StatusOK, stream)
}

// Stop ends a live stream owned by the authenticated broadcaster.
func (h *StreamHandler) Stop(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	h.controlMu.Lock()
	defer h.controlMu.Unlock()

	streamID := c.Param("id")
	sessionID := c.GetHeader("X-Stream-Session-ID")
	if err := h.useCase.End(c.Request.Context(), streamID, claims.UserID, sessionID); err != nil {
		middleware.Logger(c).Warn().Err(err).Str("stream_id", streamID).Msg("stop stream failed")
		mapStreamError(c, err)
		return
	}
	if h.hub != nil {
		// A mismatched session means a newer restart already owns the Hub. It
		// must remain connected even if this older Stop completed late.
		_ = h.hub.CloseStreamSession(streamID, claims.UserID, sessionID)
	}
	c.Status(http.StatusNoContent)
}

// Delete permanently removes a reusable stream owned by the broadcaster.
func (h *StreamHandler) Delete(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	h.controlMu.Lock()
	defer h.controlMu.Unlock()
	streamID := c.Param("id")
	if err := h.useCase.Delete(c.Request.Context(), streamID, claims.UserID); err != nil {
		mapStreamError(c, err)
		return
	}
	if h.hub != nil {
		h.hub.CloseStream(streamID)
	}
	c.Status(http.StatusNoContent)
}

// Listen records a logical join. The GET variant is StreamAudio and represents
// the real long-lived listener connection.
func (h *StreamHandler) Listen(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	streamID := c.Param("id")
	if err := h.useCase.Join(c.Request.Context(), streamID, claims.UserID); err != nil {
		middleware.Logger(c).Warn().Err(err).Str("stream_id", streamID).Msg("listen stream failed")
		mapStreamError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stream_id": streamID, "status": "listening"})
}

// StreamAudio sends publisher bytes to one authenticated listener using HTTP
// chunked transfer encoding. Request cancellation, stream stop and server
// shutdown all release the Hub registration and database listener count.
func (h *StreamHandler) StreamAudio(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
		return
	}

	streamID := c.Param("id")
	contentType, publishing := h.hub.ContentType(streamID)
	if !publishing {
		c.JSON(http.StatusConflict, gin.H{"error": "audio source is not connected"})
		return
	}
	if err := h.useCase.Join(c.Request.Context(), streamID, claims.UserID); err != nil {
		mapStreamError(c, err)
		return
	}

	client := &streaming.Client{
		ID:       uuid.NewString(),
		UserID:   claims.UserID,
		StreamID: streamID,
		Send:     make(chan []byte, h.clientBuffer),
	}
	initSegment, err := h.hub.RegisterWithInit(client)
	if err != nil {
		h.leaveDetached(streamID, claims.UserID)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		h.hub.Unregister(client)
		h.leaveDetached(streamID, claims.UserID)
	}()
	currentType, active := h.hub.ContentType(streamID)
	if !active {
		c.JSON(http.StatusConflict, gin.H{"error": "audio source disconnected"})
		return
	}
	contentType = currentType

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-store, no-transform")
	c.Header("X-Accel-Buffering", "no")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	c.Writer.Flush()
	if len(initSegment) > 0 {
		written, writeErr := c.Writer.Write(initSegment)
		if written > 0 {
			telemetry.AudioEgressBytesTotal.WithLabelValues(streamID).Add(float64(written))
			telemetry.AudioChunksTotal.WithLabelValues(streamID, "egress").Inc()
		}
		if writeErr != nil {
			return
		}
		c.Writer.Flush()
	}

	idle := time.NewTimer(h.idleTimeout)
	defer idle.Stop()
	controller := http.NewResponseController(c.Writer)
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case packet, open := <-client.Send:
			if !open {
				return
			}
			if !idle.Stop() {
				select {
				case <-idle.C:
				default:
				}
			}
			idle.Reset(h.idleTimeout)
			if h.writeTimeout > 0 {
				_ = controller.SetWriteDeadline(time.Now().Add(h.writeTimeout))
			}
			written, err := c.Writer.Write(packet)
			if written > 0 {
				telemetry.AudioEgressBytesTotal.WithLabelValues(streamID).Add(float64(written))
				telemetry.AudioChunksTotal.WithLabelValues(streamID, "egress").Inc()
			}
			if err != nil {
				middleware.Logger(c).Debug().Err(err).Str("stream_id", streamID).Msg("audio listener write ended")
				return
			}
			c.Writer.Flush()
		case <-idle.C:
			middleware.Logger(c).Warn().Str("stream_id", streamID).Msg("audio listener idle timeout")
			return
		}
	}
}

// IngestAudio receives an audio byte stream from the stream owner and fans out
// each bounded chunk through the Hub. One publisher is allowed per stream.
func (h *StreamHandler) IngestAudio(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
		return
	}
	streamID := c.Param("id")
	sessionID := c.GetHeader("X-Stream-Session-ID")
	if err := h.useCase.CanBroadcast(c.Request.Context(), streamID, claims.UserID, sessionID); err != nil {
		mapStreamError(c, err)
		return
	}
	if err := h.hub.AuthorizeStreamSession(streamID, claims.UserID, sessionID); err != nil {
		mapStreamSessionError(c, err)
		return
	}

	contentType, err := parseAudioContentType(c.GetHeader("Content-Type"))
	if err != nil {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be audio/* or application/octet-stream"})
		return
	}
	publisherCtx, err := h.hub.OpenOwnedPublisher(streamID, claims.UserID, sessionID, contentType)
	if err != nil {
		if errors.Is(err, streaming.ErrPublisherActive) || errors.Is(err, streaming.ErrPublisherNotActive) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		}
		return
	}
	defer h.hub.ClosePublisher(streamID)

	ctx, cancel := context.WithTimeout(c.Request.Context(), h.maxDuration)
	defer cancel()
	stopPublisher := context.AfterFunc(publisherCtx, cancel)
	defer stopPublisher()

	body := http.MaxBytesReader(c.Writer, c.Request.Body, h.maxIngestSize)
	defer body.Close()
	stopBody := context.AfterFunc(ctx, func() { _ = body.Close() })
	defer stopBody()

	controller := http.NewResponseController(c.Writer)
	buffer := make([]byte, h.chunkSize)
	for {
		readDeadline := time.Now().Add(h.idleTimeout)
		if deadline, ok := ctx.Deadline(); ok && deadline.Before(readDeadline) {
			readDeadline = deadline
		}
		_ = controller.SetReadDeadline(readDeadline)
		n, readErr := body.Read(buffer)
		if n > 0 {
			// Cache the first chunk (WebM EBML header) so late-joining listeners
			// on the public /audio path can initialise their browser decoder.
			h.hub.BroadcastWithInit(streamID, buffer[:n])
		}
		if readErr == nil {
			continue
		}
		if errors.Is(readErr, io.EOF) {
			c.Status(http.StatusNoContent)
			return
		}
		if ctx.Err() != nil {
			// The client may already be gone; do not try to write an error body.
			middleware.Logger(c).Info().Err(ctx.Err()).Str("stream_id", streamID).Msg("audio ingestion cancelled")
			return
		}
		if strings.Contains(readErr.Error(), "request body too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "audio ingestion byte limit exceeded"})
			return
		}
		middleware.Logger(c).Warn().Err(readErr).Str("stream_id", streamID).Msg("audio ingestion ended")
		c.JSON(http.StatusRequestTimeout, gin.H{"error": "audio ingestion idle timeout"})
		return
	}
}

// PushAudio accepts one bounded MediaRecorder blob. The first request opens a
// chunked publisher session; later requests reuse it until the stream is
// stopped. This is the browser-compatible counterpart to the continuous PUT
// endpoint, since MediaRecorder exposes ordered blobs rather than a writable
// HTTP request body.
func (h *StreamHandler) PushAudio(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
		return
	}

	streamID := c.Param("id")
	sessionID := c.GetHeader("X-Stream-Session-ID")
	contentType, err := parseAudioContentType(c.GetHeader("Content-Type"))
	if err != nil {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
		return
	}
	publisherOpen := h.hub.OwnsChunkPublisher(streamID, claims.UserID, sessionID, contentType)
	if !publisherOpen {
		// Authorize the publisher once. Later blobs are tied to the same verified
		// JWT owner and use only the Hub's in-memory session.
		if err := h.useCase.CanBroadcast(c.Request.Context(), streamID, claims.UserID, sessionID); err != nil {
			mapStreamError(c, err)
			return
		}
		if err := h.hub.AuthorizeStreamSession(streamID, claims.UserID, sessionID); err != nil {
			mapStreamSessionError(c, err)
			return
		}
	}

	limit := h.maxIngestSize
	if limit <= 0 || limit > maxBrowserAudioChunkSize {
		limit = maxBrowserAudioChunkSize
	}
	body := http.MaxBytesReader(c.Writer, c.Request.Body, limit)
	defer body.Close()
	payload, err := io.ReadAll(body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "audio chunk byte limit exceeded"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid audio chunk"})
		return
	}
	if len(payload) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio chunk is empty"})
		return
	}

	if !publisherOpen {
		err = h.hub.OpenOwnedChunkPublisher(streamID, claims.UserID, sessionID, contentType)
	}
	if err != nil {
		if errors.Is(err, streaming.ErrPublisherActive) || errors.Is(err, streaming.ErrPublisherFormatChanged) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		}
		return
	}
	if _, _, err = h.hub.BroadcastOwnedChunk(streamID, claims.UserID, sessionID, contentType, payload); err != nil {
		if errors.Is(err, streaming.ErrPublisherNotActive) || errors.Is(err, streaming.ErrPublisherFormatChanged) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *StreamHandler) leaveDetached(streamID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), h.leaveTimeout)
	defer cancel()
	if err := h.useCase.Leave(ctx, streamID, userID); err != nil {
		// The real connection has already been released. Persistence is
		// best-effort during shutdown or a database outage.
		return
	}
}

// Leave records that the authenticated listener left a stream.
func (h *StreamHandler) Leave(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	streamID := c.Param("id")
	if err := h.useCase.Leave(c.Request.Context(), streamID, claims.UserID); err != nil {
		middleware.Logger(c).Warn().Err(err).Str("stream_id", streamID).Msg("leave stream failed")
		mapStreamError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"stream_id": streamID, "status": "left"})
}

// Audio streams audio data to a listener using chunked HTTP transfer.
// The endpoint is intentionally public: the browser <audio> element cannot
// set custom Authorization headers, so auth is handled on the ingest side.
func (h *StreamHandler) Audio(c *gin.Context) {
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "audio streaming unavailable"})
		return
	}
	streamID := c.Param("id")

	// A live stream is advertised before the browser has necessarily delivered
	// its first MediaRecorder blob. Wait briefly so the response uses the exact
	// publisher MIME type instead of committing an incorrect default header.
	var contentType string
	waitForPublisher := time.NewTimer(h.idleTimeout)
	pollPublisher := time.NewTicker(25 * time.Millisecond)
	defer waitForPublisher.Stop()
	defer pollPublisher.Stop()
	for {
		if currentType, active := h.hub.ContentType(streamID); active {
			contentType = currentType
			break
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-waitForPublisher.C:
			c.JSON(http.StatusConflict, gin.H{"error": "audio source is not connected"})
			return
		case <-pollPublisher.C:
		}
	}

	connID := uuid.NewString()
	if claims, ok := middleware.GetClaims(c); ok && claims != nil {
		connID = claims.UserID
	}

	client := &streaming.Client{
		ID:       uuid.NewString(),
		UserID:   connID,
		StreamID: streamID,
		Send:     make(chan []byte, h.clientBuffer),
	}
	initSegment, err := h.hub.RegisterWithInit(client)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer h.hub.Unregister(client)
	currentType, active := h.hub.ContentType(streamID)
	if !active {
		c.JSON(http.StatusConflict, gin.H{"error": "audio source disconnected"})
		return
	}
	contentType = currentType

	// Prefer http.Flusher for real-time streaming; fall back to a no-op so the
	// handler keeps working even when a middleware wraps the ResponseWriter with
	// a type that doesn't forward the Flusher interface.
	flush := func() {}
	if f, ok := c.Writer.(http.Flusher); ok {
		flush = f.Flush
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache, no-store")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flush()

	// Send cached init segment immediately so the browser can initialise its
	// decoder even when the listener joins after the stream has started.
	if len(initSegment) > 0 {
		written, writeErr := c.Writer.Write(initSegment)
		if written > 0 {
			telemetry.AudioEgressBytesTotal.WithLabelValues(streamID).Add(float64(written))
			telemetry.AudioChunksTotal.WithLabelValues(streamID, "egress").Inc()
		}
		if writeErr != nil {
			return
		}
		flush()
	}

	middleware.Logger(c).Info().
		Str("stream_id", streamID).
		Str("conn_id", connID).
		Msg("listener connected")

	ctx := c.Request.Context()
	idle := time.NewTimer(h.idleTimeout)
	defer idle.Stop()
	controller := http.NewResponseController(c.Writer)
	for {
		select {
		case chunk, open := <-client.Send:
			if !open {
				return
			}
			if !idle.Stop() {
				select {
				case <-idle.C:
				default:
				}
			}
			idle.Reset(h.idleTimeout)
			if h.writeTimeout > 0 {
				_ = controller.SetWriteDeadline(time.Now().Add(h.writeTimeout))
			}
			written, err := c.Writer.Write(chunk)
			if written > 0 {
				telemetry.AudioEgressBytesTotal.WithLabelValues(streamID).Add(float64(written))
				telemetry.AudioChunksTotal.WithLabelValues(streamID, "egress").Inc()
			}
			if err != nil {
				return
			}
			flush()
		case <-idle.C:
			return
		case <-ctx.Done():
			middleware.Logger(c).Info().
				Str("stream_id", streamID).
				Str("conn_id", connID).
				Msg("listener disconnected")
			return
		}
	}
}
