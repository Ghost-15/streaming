package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"

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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/auth/login", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/auth/login", nil)
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
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/streams", nil)
		req.Header.Set("X-Test-User", "user-1")
		engine.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/streams", nil)
	req.Header.Set("X-Test-User", "user-1")
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}

	otherUser := httptest.NewRecorder()
	otherReq := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/streams", nil)
	otherReq.Header.Set("X-Test-User", "user-2")
	engine.ServeHTTP(otherUser, otherReq)

	if otherUser.Code != http.StatusOK {
		t.Fatalf("other user status = %d, want 200", otherUser.Code)
	}
}

func TestUserRateLimitMiddleware_FallsBackToIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.GET("/api/v1/streams",
		UserRateLimitMiddleware(60, 1),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/streams", nil)
	req.RemoteAddr = "203.0.113.55:1000"
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("first request status = %d, want 204", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/streams", nil)
	req.RemoteAddr = "203.0.113.55:1000"
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", rec.Code)
	}
}

func TestStreamDataRateLimit_AllowsMediaRecorderCadence(t *testing.T) {
	limiter := newKeyedLimiter(
		rate.Limit(streamDataRequestsPerMinute/60.0),
		streamDataBurst,
	).get("broadcaster-1")
	startedAt := time.Now()

	// Ten minutes at the Web client's two POSTs per second must not exhaust the
	// media bucket. The generic 100 req/min bucket would fail this cadence.
	for chunk := 0; chunk < 1200; chunk++ {
		at := startedAt.Add(time.Duration(chunk) * 500 * time.Millisecond)
		if !limiter.AllowN(at, 1) {
			t.Fatalf("audio chunk %d was unexpectedly rate limited", chunk+1)
		}
	}
}

func TestStreamDataRateLimitMiddleware_Returns429ByAuthenticatedUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.POST("/api/v1/streams/s1/push",
		func(c *gin.Context) {
			c.Set(claimsKey, &entity.JWTClaims{UserID: "broadcaster-1"})
			c.Next()
		},
		StreamDataRateLimitMiddleware(),
		func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		},
	)

	for i := 0; i < streamDataBurst; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/streams/s1/push", nil)
		engine.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("request %d status = %d, want 204", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/api/v1/streams/s1/push", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
}
