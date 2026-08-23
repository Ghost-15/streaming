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
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/config"
	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/router"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func writePublicKey(t *testing.T) string {
	t.Helper()
	_, path := writeKeyPair(t)
	return path
}

// writeKeyPair generates a throwaway RSA pair, writes the public half where the
// router expects it, and hands the private half back so a test can mint tokens
// the running router will actually accept.
func writeKeyPair(t *testing.T) (*rsa.PrivateKey, string) {
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
	return key, f.Name()
}

// newTestRouter builds the production router over mock repositories. Both the
// behavioural tests and the OpenAPI contract tests go through it, so there is a
// single place where the route table is assembled.
func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	engine, _ := buildTestRouter(t, true)
	return engine
}

func newTestRouterWithSwagger(t *testing.T, swaggerEnabled bool) *gin.Engine {
	t.Helper()
	engine, _ := buildTestRouter(t, swaggerEnabled)
	return engine
}

// buildTestRouter returns the assembled engine together with the signing key,
// so security tests can forge tokens accepted by this exact router.
func buildTestRouter(t *testing.T, swaggerEnabled bool) (*gin.Engine, *rsa.PrivateKey) {
	t.Helper()
	privateKey, pubPath := writeKeyPair(t)

	cfg := &config.Config{
		JWTPublicKeyPath:   pubPath,
		CORSOrigins:        "http://localhost:3000",
		MetricsBearerToken: "secret-token",
		SwaggerEnabled:     swaggerEnabled,
	}

	streamRepo := &mock.MockStreamRepository{
		ListActiveFn: func(_ context.Context) ([]entity.Stream, error) { return []entity.Stream{}, nil },
	}

	authH := handler.NewAuthHandler(usecase.NewAuthUseCase(&mock.MockUserRepository{}, &mock.MockRefreshTokenRepository{}, "", time.Hour, 720*time.Hour))
	streamH := handler.NewStreamHandler(usecase.NewStreamUseCase(streamRepo, nil))
	playlistH := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(&mock.MockPlaylistRepository{}))
	adminH := handler.NewAdminHandler(usecase.NewAdminUseCase(&mock.MockAdminRepository{}))
	favoriteH := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(&mock.MockFavoriteRepository{}))
	recommendationH := handler.NewRecommendationHandler(usecase.NewRecommendationUseCase(&mock.MockRecommendationRepository{}))

	return router.NewRouter(cfg, authH, streamH, playlistH, adminH, favoriteH, recommendationH), privateKey
}

func TestNewRouter(t *testing.T) {
	engine := newTestRouter(t)

	t.Run("health ok", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("/health status = %d, want 200", w.Code)
		}
	})

	t.Run("public streams accessible without token", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/streams", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("/api/v1/streams status = %d, want 200", w.Code)
		}
	})

	t.Run("protected route rejects missing token", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/playlists", nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("/api/v1/playlists status = %d, want 401", w.Code)
		}
	})

	t.Run("browser audio push route is protected", func(t *testing.T) {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/streams/s1/push", nil)
		req.Header.Set("Content-Type", "audio/webm;codecs=opus")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("/api/v1/streams/s1/push status = %d, want 401", w.Code)
		}
	})
}
