package middleware

import "github.com/gin-gonic/gin"

const contentSecurityPolicy = "default-src 'self'; frame-ancestors 'none'; object-src 'none'; base-uri 'self'"

// SecurityHeadersMiddleware adds defense-in-depth HTTP response headers.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		h.Set("X-Frame-Options", "DENY")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-XSS-Protection", "1; mode=block")
		h.Set("Content-Security-Policy", contentSecurityPolicy)

		c.Next()
	}
}
