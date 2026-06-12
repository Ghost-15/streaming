package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestRateLimitMiddleware_Returns429ByIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/api/v1/auth/login", RateLimitMiddleware(5, 5), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	for i := 0; i < 5; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/login", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if rec.Body.String() != "{\"error\":\"rate limit exceeded\"}" {
		t.Fatalf("body = %q, want JSON rate limit error", rec.Body.String())
	}
}

func TestUserRateLimitMiddleware_Returns429ByUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/api/v1/streams",
		func(c *gin.Context) {
			c.Set(claimsKey, &entity.JWTClaims{UserID: c.GetHeader("X-Test-User")})
			c.Next()
		},
		UserRateLimitMiddleware(100, 100),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		},
	)

	for i := 0; i < 100; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/streams", nil)
		req.Header.Set("X-Test-User", "user-1")
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/streams", nil)
	req.Header.Set("X-Test-User", "user-1")
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}

	otherUser := httptest.NewRecorder()
	otherReq := httptest.NewRequest(http.MethodGet, "/api/v1/streams", nil)
	otherReq.Header.Set("X-Test-User", "user-2")
	engine.ServeHTTP(otherUser, otherReq)

	if otherUser.Code != http.StatusOK {
		t.Fatalf("other user status = %d, want 200", otherUser.Code)
	}
}
