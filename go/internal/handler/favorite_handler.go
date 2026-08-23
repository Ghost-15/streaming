package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/usecase"
)

type FavoriteHandler struct {
	useCase usecase.FavoriteUseCase
}

func NewFavoriteHandler(uc usecase.FavoriteUseCase) *FavoriteHandler {
	return &FavoriteHandler{useCase: uc}
}

type AddFavoriteRequest struct {
	StreamID string `json:"stream_id" binding:"required"`
}

func (h *FavoriteHandler) List(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	tracks, err := h.useCase.List(c.Request.Context(), uid)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("list favorites failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Int("count", len(tracks)).Msg("listed favorites")
	c.JSON(http.StatusOK, gin.H{"data": tracks})
}

func (h *FavoriteHandler) Add(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	var req AddFavoriteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid add favorite payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.useCase.Add(c.Request.Context(), uid, req.StreamID); err != nil {
		middleware.Logger(c).Error().Err(err).Msg("add favorite failed")
		if errors.Is(err, usecase.ErrFavoriteInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Str("stream_id", req.StreamID).Msg("favorite added")
	c.Status(http.StatusCreated)
}

func (h *FavoriteHandler) Remove(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	if err := h.useCase.Remove(c.Request.Context(), uid, c.Param("streamID")); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("remove favorite failed")
		if errors.Is(err, usecase.ErrFavoriteInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "favorite not found"})
		return
	}
	middleware.Logger(c).Info().Str("stream_id", c.Param("streamID")).Msg("favorite removed")
	c.Status(http.StatusNoContent)
}
