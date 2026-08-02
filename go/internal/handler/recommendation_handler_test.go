package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func recoEngine(h *handler.RecommendationHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if userID != "" {
		r.Use(func(c *gin.Context) {
			c.Set("claims", &entity.JWTClaims{UserID: userID, Role: entity.RoleUser})
			c.Next()
		})
	}
	r.GET("/recommendations", h.List)
	return r
}

func TestRecommendationHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockRecommendationRepository{
			RecommendStreamsFn: func(_ context.Context, _ string, _ int) ([]entity.Stream, error) {
				return []entity.Stream{{ID: "s1", Title: "Popular"}}, nil
			},
		}
		h := handler.NewRecommendationHandler(usecase.NewRecommendationUseCase(repo))
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/recommendations", nil)
		w := httptest.NewRecorder()
		recoEngine(h, "u1").ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("List status = %d, want 200", w.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h := handler.NewRecommendationHandler(usecase.NewRecommendationUseCase(&mock.MockRecommendationRepository{}))
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/recommendations", nil)
		w := httptest.NewRecorder()
		recoEngine(h, "").ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("List status = %d, want 401", w.Code)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockRecommendationRepository{
			RecommendStreamsFn: func(_ context.Context, _ string, _ int) ([]entity.Stream, error) { return nil, errStreamTest },
		}
		h := handler.NewRecommendationHandler(usecase.NewRecommendationUseCase(repo))
		req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/recommendations", nil)
		w := httptest.NewRecorder()
		recoEngine(h, "u1").ServeHTTP(w, req)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("List status = %d, want 500", w.Code)
		}
	})
}
