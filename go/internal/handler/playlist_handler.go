package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

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

// CreateRequest is the JSON body for POST /playlists.
type CreatePlaylistRequest struct {
	Title string `json:"title" binding:"required,min=1,max=100"`
}

// List godoc.
// @Summary     List playlists owned by the authenticated user
// @Tags        playlists
// @Produce     json
// @Success     200 {array} entity.Playlist
// @Router      /api/v1/playlists [get]
func (h *PlaylistHandler) List(c *gin.Context) {
	// TODO Sprint 2 — US-007: extract ownerID from JWT claims
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

// Create godoc
// @Summary     Create a new playlist
// @Tags        playlists
// @Accept      json
// @Produce     json
// @Param       body body CreatePlaylistRequest true "Playlist payload"
// @Success     201 {object} entity.Playlist
// @Router      /api/v1/playlists [post]
func (h *PlaylistHandler) Create(c *gin.Context) {
	var req CreatePlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// TODO Sprint 2 — US-007
	c.JSON(http.StatusCreated, gin.H{"title": req.Title})
}

// Delete godoc
// @Summary     Delete a playlist
// @Tags        playlists
// @Param       id path string true "Playlist ID"
// @Success     204
// @Router      /api/v1/playlists/{id} [delete]
func (h *PlaylistHandler) Delete(c *gin.Context) {
	// TODO Sprint 2 — US-007
	c.Status(http.StatusNoContent)
}
