package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// StreamHandler handles HTTP requests for live streams.
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
