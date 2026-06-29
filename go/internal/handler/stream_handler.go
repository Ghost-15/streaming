package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/infrastructure/streaming"
	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// StreamHandler handles HTTP requests for live streams.
// Sprint 1 — US-003, US-007.
type StreamHandler struct {
	useCase usecase.StreamUseCase
	hub     *streaming.Hub
}

// NewStreamHandler creates a new StreamHandler wired with the streaming Hub.
func NewStreamHandler(uc usecase.StreamUseCase, hub *streaming.Hub) *StreamHandler {
	return &StreamHandler{useCase: uc, hub: hub}
}

// StartRequest is the JSON body for POST /streams.
// stream_url is optional: if omitted, a demo audio source is used so the stream is
// immediately playable end-to-end (real broadcaster ingestion is out of scope — ADR-007).
type StartRequest struct {
	Title     string `json:"title"      binding:"required,min=3,max=100"`
	StreamURL string `json:"stream_url" binding:"omitempty,url"`
}

// defaultStreamURL is a royalty-free audio source used when a broadcaster
// does not provide a stream_url (keeps the demo playable end-to-end).
const defaultStreamURL = "https://www.soundhelix.com/examples/mp3/SoundHelix-Song-1.mp3"

// mapStreamError converts sentinel usecase errors to HTTP responses.
func mapStreamError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrStreamNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "stream not found"})
	case errors.Is(err, usecase.ErrStreamForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "not the broadcaster"})
	case errors.Is(err, usecase.ErrStreamEnded):
		c.JSON(http.StatusConflict, gin.H{"error": "stream already ended"})
	case errors.Is(err, usecase.ErrStreamInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// ListActive godoc
// @Summary     List all live streams
// @Tags        streams
// @Produce     json
// @Success     200 {array} entity.Stream
// @Router      /api/v1/streams [get]
func (h *StreamHandler) ListActive(c *gin.Context) {
	streams, err := h.useCase.ListActive(c.Request.Context())
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("list active streams failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Int("count", len(streams)).Msg("listed active streams")
	c.JSON(http.StatusOK, gin.H{"data": streams})
}

// Start godoc
// @Summary     Start a new live stream (diffuseur role required)
// @Tags        streams
// @Accept      json
// @Produce     json
// @Param       body body StartRequest true "Stream payload"
// @Success     201 {object} entity.Stream
// @Router      /api/v1/streams [post]
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

	streamURL := req.StreamURL
	if streamURL == "" {
		streamURL = defaultStreamURL
	}
	stream, err := h.useCase.Start(c.Request.Context(), claims.UserID, req.Title, streamURL)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("start stream failed")
		mapStreamError(c, err)
		return
	}

	// US-010 metrics: a new stream went live.
	telemetry.StreamStartTotal.Inc()
	telemetry.ActiveStreams.Inc()

	middleware.Logger(c).Info().Str("stream_id", stream.ID).Msg("stream started")
	c.JSON(http.StatusCreated, stream)
}

// Stop godoc
// @Summary     Stop a live stream (broadcaster only)
// @Tags        streams
// @Param       id path string true "Stream ID"
// @Success     204
// @Router      /api/v1/streams/{id}/stop [put]
func (h *StreamHandler) Stop(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	if err := h.useCase.End(c.Request.Context(), c.Param("id"), claims.UserID); err != nil {
		middleware.Logger(c).Error().Err(err).Msg("stop stream failed")
		mapStreamError(c, err)
		return
	}

	telemetry.ActiveStreams.Dec()
	middleware.Logger(c).Info().Str("stream_id", c.Param("id")).Msg("stream stopped")
	c.Status(http.StatusNoContent)
}

// Join godoc
// @Summary     Join a stream as a listener (returns the stream with its audio URL)
// @Tags        streams
// @Produce     json
// @Param       id path string true "Stream ID"
// @Success     200 {object} entity.Stream
// @Router      /api/v1/streams/{id}/join [post]
func (h *StreamHandler) Join(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	stream, err := h.useCase.Join(c.Request.Context(), c.Param("id"), claims.UserID)
	if err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("join stream failed")
		mapStreamError(c, err)
		return
	}
	middleware.Logger(c).Info().Str("stream_id", stream.ID).Msg("listener joined")
	c.JSON(http.StatusOK, stream)
}

// Leave godoc
// @Summary     Leave a stream
// @Tags        streams
// @Param       id path string true "Stream ID"
// @Success     204
// @Router      /api/v1/streams/{id}/leave [post]
func (h *StreamHandler) Leave(c *gin.Context) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	if err := h.useCase.Leave(c.Request.Context(), c.Param("id"), claims.UserID); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("leave stream failed")
		mapStreamError(c, err)
		return
	}
	middleware.Logger(c).Info().Str("stream_id", c.Param("id")).Msg("listener left")
	c.Status(http.StatusNoContent)
}

// Listen opens a Server-Sent Events channel for real-time stream events.
// The Hub tracks the listener (feeding the listeners_per_stream metric) and pushes
// periodic listener-count updates. Audio itself is played client-side from stream_url.
// Sprint 1 — US-003.
func (h *StreamHandler) Listen(c *gin.Context) {
	streamID := c.Param("id")
	userID := "anonymous"
	if claims, ok := middleware.GetClaims(c); ok && claims != nil {
		userID = claims.UserID
	}

	client := &streaming.Client{
		UserID:   userID,
		StreamID: streamID,
		Send:     make(chan []byte, 16),
	}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial event so the client knows it is connected.
	c.SSEvent("connected", gin.H{"stream_id": streamID, "listeners": h.hub.ListenerCount(streamID)})
	c.Writer.Flush()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case msg, ok := <-client.Send:
			if !ok {
				return
			}
			c.SSEvent("message", string(msg))
			c.Writer.Flush()
		case <-ticker.C:
			c.SSEvent("listeners", gin.H{"count": h.hub.ListenerCount(streamID)})
			c.Writer.Flush()
		}
	}
}
