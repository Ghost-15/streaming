package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// StreamHandler handles HTTP requests for live streams.
type StreamHandler struct {
	useCase usecase.StreamUseCase
	hub     *streaming.Hub
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(uc usecase.StreamUseCase, hub *streaming.Hub) *StreamHandler {
	return &StreamHandler{useCase: uc, hub: hub}
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

	if err := h.useCase.End(c.Request.Context(), c.Param("id"), claims.UserID); err != nil {
		middleware.Logger(c).Warn().Err(err).Str("stream_id", c.Param("id")).Msg("stop stream failed")
		mapStreamError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Listen records that the authenticated listener joined a stream.
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

// Push receives audio chunks from the broadcaster (one POST per chunk) and
// fans them out to all connected listeners via the Hub.
func (h *StreamHandler) Push(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	streamID := c.Param("id")
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "audio/aac"
	}
	h.hub.SetContentType(streamID, contentType)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(body) == 0 {
		c.Status(http.StatusOK)
		return
	}

	// Cache the first chunk: it contains the WebM EBML header that late-joining
	// listeners need to initialise their browser decoder.
	h.hub.SetInitSegment(streamID, body)
	h.hub.Broadcast(streamID, body)

	middleware.Logger(c).Debug().
		Str("stream_id", streamID).
		Str("user_id", claims.UserID).
		Int("bytes", len(body)).
		Msg("audio chunk pushed")

	c.Status(http.StatusOK)
}

// Audio streams audio data to a listener using chunked HTTP transfer.
// The endpoint is intentionally public: the browser <audio> element cannot
// set custom Authorization headers, so auth is handled at the Push side.
func (h *StreamHandler) Audio(c *gin.Context) {
	streamID := c.Param("id")

	connID := uuid.NewString()
	if claims, ok := middleware.GetClaims(c); ok && claims != nil {
		connID = claims.UserID
	}

	client := &streaming.Client{
		UserID:   connID,
		StreamID: streamID,
		Send:     make(chan []byte, 128),
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// Prefer http.Flusher for real-time streaming; fall back to a no-op so the
	// handler keeps working even when a middleware wraps the ResponseWriter with
	// a type that doesn't forward the Flusher interface.
	flush := func() {}
	if f, ok := c.Writer.(http.Flusher); ok {
		flush = f.Flush
	}

	c.Header("Content-Type", h.hub.ContentType(streamID))
	c.Header("Cache-Control", "no-cache, no-store")
	c.Header("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flush()

	// Send cached init segment immediately so the browser can initialise its
	// decoder even when the listener joins after the stream has started.
	if init := h.hub.InitSegment(streamID); len(init) > 0 {
		_, _ = c.Writer.Write(init)
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
