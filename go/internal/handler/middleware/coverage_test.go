package middleware_test

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler/middleware"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.CORSMiddleware("http://localhost:3000, https://streampulse.app"))
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	// Preflight request exercises the CORS logic.
	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("CORS allow-origin = %q, want http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.MetricsMiddleware())
	r.GET("/ping", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("MetricsMiddleware passthrough status = %d, want 200", w.Code)
	}
}

func signToken(t *testing.T, key *rsa.PrivateKey, role entity.UserRole) string {
	t.Helper()
	claims := entity.JWTClaims{
		UserID: "u1",
		Email:  "u1@test.com",
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return tok
}

func TestRBACMiddleware_ValidAndForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen key: %v", err)
	}
	pub := &key.PublicKey

	newEngine := func(allowed ...entity.UserRole) *gin.Engine {
		r := gin.New()
		r.GET("/protected", middleware.RBACMiddleware(pub, allowed...), func(c *gin.Context) {
			claims, ok := middleware.GetClaims(c)
			if !ok || claims == nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			c.String(http.StatusOK, string(claims.Role))
		})
		return r
	}

	t.Run("valid role passes", func(t *testing.T) {
		r := newEngine(entity.RoleUser, entity.RoleAdmin)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+signToken(t, key, entity.RoleUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("valid role status = %d, want 200", w.Code)
		}
	})

	t.Run("wrong role forbidden", func(t *testing.T) {
		r := newEngine(entity.RoleAdmin)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+signToken(t, key, entity.RoleUser))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("wrong role status = %d, want 403", w.Code)
		}
	})

	t.Run("missing token", func(t *testing.T) {
		r := newEngine(entity.RoleUser)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("missing token status = %d, want 401", w.Code)
		}
	})

	t.Run("invalid bearer format", func(t *testing.T) {
		r := newEngine(entity.RoleUser)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "not-a-bearer-token")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("invalid bearer status = %d, want 401", w.Code)
		}
	})

	t.Run("malformed token", func(t *testing.T) {
		r := newEngine(entity.RoleUser)
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("malformed token status = %d, want 401", w.Code)
		}
	})
}

func TestGetClaimsWithoutMiddleware(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	claims, ok := middleware.GetClaims(c)
	if ok || claims != nil {
		t.Fatalf("GetClaims() = (%v, %t), want (nil, false)", claims, ok)
	}
}
