package middleware

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"os"
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

// loadPublicKey reads and parses the RSA public key from a PEM file.
func loadPublicKey(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return pubKey, nil
}

// roleOrdinal defines the role precedence: Anon < User < Diffuseur < Admin
// Higher ordinal = more permissions.
func roleOrdinal(r entity.UserRole) int {
	switch r {
	case entity.RoleAnon:
		return 0
	case entity.RoleUser:
		return 1
	case entity.RoleDiffuseur:
		return 2
	case entity.RoleAdmin:
		return 3
	default:
		return -1 // Unknown role
	}
}

// RBACMiddleware validates the JWT RS256 token and checks that the caller's
// role is in the allowed set. Returns 401 if missing/invalid, 403 if forbidden.
func RBACMiddleware(allowedRoles ...entity.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// STR-15: Extract Bearer token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// STR-15: Parse and validate JWT RS256 with public key
		claims := &JWTClaims{}
		pubKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
		if pubKeyPath == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "public key path not configured"})
			return
		}

		pubKey, err := loadPublicKey(pubKeyPath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to load public key"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Verify the alg is RS256
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return pubKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// STR-16: Check role against allowedRoles (role hierarchy)
		userRoleOrdinal := roleOrdinal(claims.Role)
		hasPermission := false
		for _, allowedRole := range allowedRoles {
			if userRoleOrdinal >= roleOrdinal(allowedRole) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		// Set claims in context for downstream handlers
		c.Set(claimsKey, claims)
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

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
