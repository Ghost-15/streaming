package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/handler"
)

func TestHealthReturnsLivenessPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", handler.Health)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Health() status = %d, want 200", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Health() body is not JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("Health() status payload = %q, want ok", body["status"])
	}
	if body["service"] != "streampulse-api" {
		t.Fatalf("Health() service payload = %q, want streampulse-api", body["service"])
	}
}
