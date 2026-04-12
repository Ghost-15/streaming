package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures strict CORS for the API.
// origins vient de config.Config.CORSOrigins (ENV CORS_ALLOWED_ORIGINS).
// Supporte plusieurs origines séparées par une virgule :
//
//	CORS_ALLOWED_ORIGINS=http://localhost:3000,https://streampulse.app
//
// Sprint 1 — US-015.
func CORSMiddleware(origins string) gin.HandlerFunc {
	parsed := parseOrigins(origins)
	cfg := cors.Config{
		AllowOrigins:     parsed,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}
	return cors.New(cfg)
}

// parseOrigins splits a comma-separated list of origins and trims whitespace.
// Retourne une slice vide si raw est vide — le caller (config.validate) garantit
// que CORS_ALLOWED_ORIGINS est défini avant d'arriver ici.
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
