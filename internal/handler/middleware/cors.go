package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures strict CORS for the API.
// AllowOrigins is read from ENV CORS_ALLOWED_ORIGINS (no wildcard in prod).
// Sprint 1 — US-015 (security hardening sprint 3, basic CORS sprint 1).
func CORSMiddleware() gin.HandlerFunc {
	cfg := cors.Config{
		// TODO Sprint 1: read from config.Config.CORSOrigins
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Authorization", "Content-Type"},
	}
	return cors.New(cfg)
}
