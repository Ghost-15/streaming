package router_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Ghost-15/streaming/internal/config"
	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/router"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func writePublicKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("gen key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal pub: %v", err)
	}
	f, err := os.CreateTemp("", "router-test-pub-*.pem")
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	if err := pem.Encode(f, &pem.Block{Type: "PUBLIC KEY", Bytes: der}); err != nil {
		t.Fatalf("encode pem: %v", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestNewRouter(t *testing.T) {
	pubPath := writePublicKey(t)

	cfg := &config.Config{
		JWTPublicKeyPath:   pubPath,
		CORSOrigins:        "http://localhost:3000",
		MetricsBearerToken: "secret-token",
	}

	streamRepo := &mock.MockStreamRepository{
		ListActiveFn: func(_ context.Context) ([]entity.Stream, error) { return []entity.Stream{}, nil },
	}

	authH := handler.NewAuthHandler(usecase.NewAuthUseCase(&mock.MockUserRepository{}, ""))
	streamH := handler.NewStreamHandler(usecase.NewStreamUseCase(streamRepo, nil))
	playlistH := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(&mock.MockPlaylistRepository{}))
	adminH := handler.NewAdminHandler(usecase.NewAdminUseCase(&mock.MockAdminRepository{}))
	favoriteH := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(&mock.MockFavoriteRepository{}))
	recommendationH := handler.NewRecommendationHandler(usecase.NewRecommendationUseCase(&mock.MockRecommendationRepository{}))

	engine := router.NewRouter(cfg, authH, streamH, playlistH, adminH, favoriteH, recommendationH)

	t.Run("health ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("/health status = %d, want 200", w.Code)
		}
	})

	t.Run("public streams accessible without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/streams", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("/api/v1/streams status = %d, want 200", w.Code)
		}
	})

	t.Run("protected route rejects missing token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/playlists", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("/api/v1/playlists status = %d, want 401", w.Code)
		}
	})
}
