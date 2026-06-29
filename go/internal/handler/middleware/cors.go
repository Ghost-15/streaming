package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures strict CORS for the API.
// origins comes from config.Config.CORSOrigins (ENV CORS_ALLOWED_ORIGINS).
// Supports multiple origins separated by a comma:
//
//	CORS_ALLOWED_ORIGINS=http://localhost:3000,https://streampulse.app
//
// Sprint 1 — US-015.
func CORSMiddleware(origins string) gin.HandlerFunc {
	parsed := parseOrigins(origins)
	cfg := cors.Config{
		AllowOrigins:     parsed,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}
	return cors.New(cfg)
}

// parseOrigins splits a comma-separated list of allowed origins and trims whitespace.
// Returns an empty slice if raw is empty — config.validate ensures
// CORS_ALLOWED_ORIGINS is set before this is called.
func parseOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
