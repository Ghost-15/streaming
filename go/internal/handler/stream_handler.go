package handler

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
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

func mapStreamError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrStreamInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	case errors.Is(err, usecase.ErrStreamNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
	case errors.Is(err, usecase.ErrStreamForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
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
	middleware.Logger(c).Info().Str("stream_id", stream.ID).Msg("stream started")
	c.JSON(http.StatusCreated, stream)
}

// Stop ends a live stream owned by the authenticated broadcaster.
func (h *StreamHandler) Stop(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	streamID := c.Param("id")
	if err := h.useCase.End(c.Request.Context(), streamID, claims.UserID); err != nil {
		middleware.Logger(c).Warn().Err(err).Str("stream_id", streamID).Msg("stop stream failed")
		mapStreamError(c, err)
		return
	}
	if h.hub != nil {
		h.hub.CloseStream(c.Param("id"))
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
	if err := h.hub.Register(client); err != nil {
		h.leaveDetached(streamID, claims.UserID)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer func() {
		h.hub.Unregister(client)
		h.leaveDetached(streamID, claims.UserID)
	}()

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-store, no-transform")
	c.Header("X-Accel-Buffering", "no")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
	c.Writer.Flush()

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
	if err := h.useCase.CanBroadcast(c.Request.Context(), streamID, claims.UserID); err != nil {
		mapStreamError(c, err)
		return
	}

	contentType := c.GetHeader("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || (!strings.HasPrefix(mediaType, "audio/") && mediaType != "application/octet-stream") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Content-Type must be audio/* or application/octet-stream"})
		return
	}
	publisherCtx, err := h.hub.OpenPublisher(streamID, mediaType)
	if err != nil {
		if errors.Is(err, streaming.ErrPublisherActive) {
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
			h.hub.SetInitSegment(streamID, buffer[:n])
			h.hub.Broadcast(streamID, buffer[:n])
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

	connID := uuid.NewString()
	if claims, ok := middleware.GetClaims(c); ok && claims != nil {
		connID = claims.UserID
	}

	client := &streaming.Client{
		ID:       uuid.NewString(),
		UserID:   connID,
		StreamID: streamID,
		Send:     make(chan []byte, 128),
	}
	if err := h.hub.Register(client); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}
	defer h.hub.Unregister(client)
	initSegment := h.hub.InitSegment(streamID)

	// Prefer http.Flusher for real-time streaming; fall back to a no-op so the
	// handler keeps working even when a middleware wraps the ResponseWriter with
	// a type that doesn't forward the Flusher interface.
	flush := func() {}
	if f, ok := c.Writer.(http.Flusher); ok {
		flush = f.Flush
	}

	contentType, ok := h.hub.ContentType(streamID)
	if !ok || contentType == "" {
		contentType = "audio/webm; codecs=opus"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache, no-store")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flush()

	// Send cached init segment immediately so the browser can initialise its
	// decoder even when the listener joins after the stream has started.
	if len(initSegment) > 0 {
		_, _ = c.Writer.Write(initSegment)
		flush()
	}

	middleware.Logger(c).Info().
		Str("stream_id", streamID).
		Str("conn_id", connID).
		Msg("listener connected")

	ctx := c.Request.Context()
	for {
		select {
		case chunk, open := <-client.Send:
			if !open {
				return
			}
			if _, err := c.Writer.Write(chunk); err != nil {
				return
			}
			flush()
		case <-ctx.Done():
			middleware.Logger(c).Info().
				Str("stream_id", streamID).
				Str("conn_id", connID).
				Msg("listener disconnected")
			return
		}
	}
}
