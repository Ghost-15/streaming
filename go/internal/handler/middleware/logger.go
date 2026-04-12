package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// ZerologMiddleware logs each request as a structured JSON line.
// Injects trace_id from OTEL context when available.
// Sprint 1 — US-006.
func ZerologMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("latency_ms", time.Since(start)).
			Str("client_ip", c.ClientIP()).
			// TODO Sprint 2 — US-008: inject trace_id from OTEL span context
			Msg("request")
	}
}
