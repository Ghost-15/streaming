package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler/middleware"
)

func TestMetricsAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		expectedToken  string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "disabled when token is empty",
			expectedToken:  "",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing token",
			expectedToken:  "secret",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong token",
			expectedToken:  "secret",
			authHeader:     "Bearer nope",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "valid bearer token",
			expectedToken:  "secret",
			authHeader:     "Bearer secret",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := gin.New()
			engine.GET("/metrics", middleware.MetricsAuthMiddleware(tt.expectedToken), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"status": "ok"})
			})

			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/metrics", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			engine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.expectedStatus)
			}
		})
	}
}
