package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// keyedLimiter stores rate limiters per caller key.
type keyedLimiter struct {
	mu       sync.Mutex
	limiters map[string]*rate.Limiter
	r        rate.Limit
	b        int
}

func newKeyedLimiter(r rate.Limit, b int) *keyedLimiter {
	return &keyedLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (kl *keyedLimiter) get(key string) *rate.Limiter {
	kl.mu.Lock()
	defer kl.mu.Unlock()
	if l, ok := kl.limiters[key]; ok {
		return l
	}
	l := rate.NewLimiter(kl.r, kl.b)
	kl.limiters[key] = l
	return l
}

// RateLimitMiddleware limits requests per IP.
func RateLimitMiddleware(reqPerMin float64, burst int) gin.HandlerFunc {
	return rateLimitByKey(reqPerMin, burst, func(c *gin.Context) string {
		return c.ClientIP()
	})
}

// UserRateLimitMiddleware limits requests per authenticated user.
func UserRateLimitMiddleware(reqPerMin float64, burst int) gin.HandlerFunc {
	return rateLimitByKey(reqPerMin, burst, func(c *gin.Context) string {
		if claims, ok := GetClaims(c); ok && claims.UserID != "" {
			return claims.UserID
		}
		return c.ClientIP()
	})
}

func rateLimitByKey(reqPerMin float64, burst int, keyFn func(*gin.Context) string) gin.HandlerFunc {
	limiter := newKeyedLimiter(rate.Limit(reqPerMin/60), burst)
	return func(c *gin.Context) {
		if !limiter.get(keyFn(c)).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
