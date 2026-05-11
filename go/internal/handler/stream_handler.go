package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// StreamHandler handles HTTP requests for live streams.
// Sprint 1 — US-003, US-007.
type StreamHandler struct {
	useCase usecase.StreamUseCase
}

// NewStreamHandler creates a new StreamHandler.
func NewStreamHandler(uc usecase.StreamUseCase) *StreamHandler {
	return &StreamHandler{useCase: uc}
}

// StartRequest is the JSON body for POST /streams.
type StartRequest struct {
	Title string `json:"title" binding:"required,min=3,max=100"`
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
	c.JSON(http.StatusOK, streams)
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
	var req StartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid start stream payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO Sprint 1 — US-003: extract broadcasterID from JWT claims
	stream, err := h.useCase.Start(c.Request.Context(), "", req.Title)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("start stream failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Str("stream_id", stream.ID).Msg("stream started")

	c.JSON(http.StatusCreated, stream)
}

// Listen serves a Server-Sent Events stream for a listener.
// Sprint 1 — US-003.
func (h *StreamHandler) Listen(c *gin.Context) {
	streamID := c.Param("id")

	// TODO Sprint 1 — US-003:
	// 1. Register client in Hub
	// 2. Set SSE headers
	// 3. Stream audio chunks until disconnect

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.JSON(http.StatusOK, gin.H{"stream_id": streamID, "status": "TODO"})
}
