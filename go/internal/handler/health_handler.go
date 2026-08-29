package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health godoc.
// @Summary     Liveness probe
// @Description Used by the Render health check and by docker compose. It answers
// @Description without touching the database, so a database outage does not take
// @Description the container down.
// @Tags        operations
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "streampulse-api"})
}
