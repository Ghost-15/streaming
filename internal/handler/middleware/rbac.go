package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Ghost-15/streaming/internal/entity"
)

const claimsKey = "claims"

// JWTClaims holds the JWT payload fields used by StreamPulse.
type JWTClaims struct {
	UserID string          `json:"sub"`
	Email  string          `json:"email"`
	Role   entity.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// RBACMiddleware validates the JWT RS256 token and checks that the caller's
// role is in the allowed set. Returns 401 if missing/invalid, 403 if forbidden.
// Sprint 1 — US-002.
func RBACMiddleware(allowedRoles ...entity.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO Sprint 1 — US-002:
		// 1. Extract Bearer token from Authorization header
		// 2. Parse and validate JWT RS256 with public key
		// 3. Check role against allowedRoles
		// 4. Set claims in context for downstream handlers

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		// Placeholder — real implementation in Sprint 1
		c.Next()
	}
}

// GetClaims retrieves the JWT claims set by RBACMiddleware from the context.
func GetClaims(c *gin.Context) (*JWTClaims, bool) {
	v, exists := c.Get(claimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := v.(*JWTClaims)
	return claims, ok
}
