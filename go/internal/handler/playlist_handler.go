package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// PlaylistHandler handles HTTP requests for playlists.
// Sprint 2 — US-007.
type PlaylistHandler struct {
	useCase usecase.PlaylistUseCase
}

// NewPlaylistHandler creates a new PlaylistHandler.
func NewPlaylistHandler(uc usecase.PlaylistUseCase) *PlaylistHandler {
	return &PlaylistHandler{useCase: uc}
}

// CreatePlaylistRequest is the JSON body for POST /playlists.
type CreatePlaylistRequest struct {
	Title string `json:"title" binding:"required,min=1,max=100"`
}

// UpdatePlaylistRequest is the JSON body for PUT /playlists/:id.
type UpdatePlaylistRequest struct {
	Title string `json:"title" binding:"required,min=1,max=100"`
}

// AddTrackRequest is the JSON body for POST /playlists/:id/tracks.
type AddTrackRequest struct {
	TrackID string `json:"track_id" binding:"required"`
}

// ownerID extracts the authenticated user ID from the request context.
// Returns false if no claims are present (middleware misconfiguration).
func ownerID(c *gin.Context) (string, bool) {
	claims, ok := middleware.GetClaims(c)
	if !ok || claims == nil {
		return "", false
	}
	return claims.UserID, true
}

// mapPlaylistError converts sentinel usecase errors to HTTP responses.
// Returns true if the error was handled (response written).
func mapPlaylistError(c *gin.Context, err error) bool {
	switch {
	case errors.Is(err, usecase.ErrPlaylistNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "playlist not found"})
	case errors.Is(err, usecase.ErrPlaylistForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, usecase.ErrPlaylistInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	return true
}

// List godoc.
// @Summary     List playlists owned by the authenticated user
// @Tags        playlists
// @Produce     json
// @Success     200 {array} entity.Playlist
// @Failure     401 {object} map[string]string
// @Router      /api/v1/playlists [get]
func (h *PlaylistHandler) List(c *gin.Context) {
	// TODO Sprint 2 — US-007: extract ownerID from JWT claims
	middleware.Logger(c).Info().Msg("listed playlists")
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// Create godoc.
// @Summary     Create a new playlist
// @Tags        playlists
// @Accept      json
// @Produce     json
// @Param       body body CreatePlaylistRequest true "Playlist payload"
// @Success     201 {object} entity.Playlist
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Router      /api/v1/playlists [post]
func (h *PlaylistHandler) Create(c *gin.Context) {
	// _ discards ownerID until Sprint 2 — US-007 implements usecase.Create(ctx, ownerID, title).
	_, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid create playlist payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO Sprint 2 — US-007
	middleware.Logger(c).Info().Str("title", req.Title).Msg("playlist created")
	c.JSON(http.StatusCreated, gin.H{"title": req.Title})
}

// Update godoc.
// @Summary     Update a playlist title
// @Tags        playlists
// @Accept      json
// @Produce     json
// @Param       id path string true "Playlist ID"
// @Param       body body UpdatePlaylistRequest true "Update payload"
// @Success     200 {object} entity.Playlist
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/v1/playlists/{id} [put]
func (h *PlaylistHandler) Update(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	var req UpdatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.Logger(c).Warn().Err(err).Msg("invalid update playlist payload")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	playlist, err := h.useCase.Update(c.Request.Context(), c.Param("id"), uid, req.Title)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("update playlist failed")
		mapPlaylistError(c, err)
		return
	}
	middleware.Logger(c).Info().Str("playlist_id", playlist.ID).Msg("playlist updated")
	c.JSON(http.StatusOK, playlist)
}

// Delete godoc.
// @Summary     Delete a playlist
// @Tags        playlists
// @Param       id path string true "Playlist ID"
// @Success     204
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Router      /api/v1/playlists/{id} [delete]
func (h *PlaylistHandler) Delete(c *gin.Context) {
	// TODO Sprint 2 — US-007
	middleware.Logger(c).Info().Str("playlist_id", c.Param("id")).Msg("playlist deleted")
	c.Status(http.StatusNoContent)
}
