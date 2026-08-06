package middleware

import (
	"fmt"
	"strings"
	"time"

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
	fmt.Printf("[CORS] allowed origins: %v\n", parsed)

	// Check if "http://localhost" (no port) is listed — acts as wildcard for all localhost ports.
	allowAllLocalhost := false
	for _, o := range parsed {
		if o == "http://localhost" {
			allowAllLocalhost = true
			break
		}
	}

	cfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Stream-Session-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			if allowAllLocalhost && strings.HasPrefix(origin, "http://localhost:") {
				fmt.Printf("[CORS] ✓ localhost wildcard: %s\n", origin)
				return true
			}
			for _, o := range parsed {
				if o == origin {
					fmt.Printf("[CORS] ✓ allowed: %s\n", origin)
					return true
				}
			}
			fmt.Printf("[CORS] ✗ blocked: %s\n", origin)
			return false
		},
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
