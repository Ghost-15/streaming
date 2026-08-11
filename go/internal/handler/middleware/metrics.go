package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/infrastructure/telemetry"
)

// MetricsMiddleware records HTTP duration for Prometheus and direct OTLP export.
// Sprint 3 — US-010. Must be registered after RouterEngine.Use(ZerologMiddleware).
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// FullPath returns the Gin route pattern (e.g. "/api/v1/playlists/:id")
		// which avoids exploding cardinality with raw IDs in the metric labels.
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		telemetry.ObserveAPIRequestDuration(
			c.Request.Context(),
			route,
			c.Request.Method,
			strconv.Itoa(c.Writer.Status()),
			time.Since(start).Seconds(),
		)
	}
}
