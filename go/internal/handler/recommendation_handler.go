package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
	"github.com/Ghost-15/streaming/internal/usecase"
)

// RecommendationHandler serves personalised stream recommendations.
type RecommendationHandler struct {
	useCase usecase.RecommendationUseCase
}

func NewRecommendationHandler(uc usecase.RecommendationUseCase) *RecommendationHandler {
	return &RecommendationHandler{useCase: uc}
}

// List godoc.
// @Summary     Personalised live-stream recommendations
// @Tags        recommendations
// @Produce     json
// @Success     200 {array} entity.Stream
// @Failure     401 {object} map[string]string
// @Router      /api/v1/recommendations [get]
func (h *RecommendationHandler) List(c *gin.Context) {
	uid, ok := ownerID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing claims"})
		return
	}

	streams, err := h.useCase.Recommend(c.Request.Context(), uid)
	if err != nil {
		middleware.Logger(c).Error().Err(err).Msg("recommendations failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	middleware.Logger(c).Info().Int("count", len(streams)).Msg("served recommendations")
	c.JSON(http.StatusOK, gin.H{"data": streams})
}
