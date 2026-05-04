package middleware

import (
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Ghost-15/streaming/internal/entity"
)

const claimsKey = "claims"

// RBACMiddleware validates the JWT RS256 token and checks that the caller's
// role is in the allowed set.
// Returns 401 if missing/invalid token, 403 if the role is not allowed.
// Sprint 1 — US-002.
func RBACMiddleware(publicKey *rsa.PublicKey, allowedRoles ...entity.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract Bearer token from Authorization header.
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse and validate the JWT RS256 signature.
		claims := &entity.JWTClaims{}
		_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return publicKey, nil
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// 3. Check the role against the allowed set.
		allowed := false
		for _, r := range allowedRoles {
			if claims.Role == r {
				allowed = true
				break
			}
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		// 4. Inject claims into the context for downstream handlers.
		c.Set(claimsKey, claims)
		c.Next()
	}
}

// GetClaims retrieves the JWT claims set by RBACMiddleware from the context.
func GetClaims(c *gin.Context) (*entity.JWTClaims, bool) {
	v, exists := c.Get(claimsKey)
	if !exists {
		return nil, false
	}
	claims, ok := v.(*entity.JWTClaims)
	return claims, ok
}
