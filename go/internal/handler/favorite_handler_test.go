package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func newFavoriteEngine(h *handler.FavoriteHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if userID != "" {
		r.Use(func(c *gin.Context) {
			c.Set("claims", &entity.JWTClaims{UserID: userID, Role: entity.RoleUser})
			c.Next()
		})
	}
	r.GET("/favorites", h.List)
	r.POST("/favorites", h.Add)
	r.DELETE("/favorites/:trackID", h.Remove)
	return r
}

func favReq(engine *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var rd *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	} else {
		rd = bytes.NewReader(nil)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, rd)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestFavoriteHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockFavoriteRepository{
			ListByUserFn: func(_ context.Context, _ string) ([]entity.Track, error) {
				return []entity.Track{{ID: "t1", Title: "Song"}}, nil
			},
		}
		h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(repo))
		w := favReq(newFavoriteEngine(h, "u1"), "GET", "/favorites", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("List status = %d, want 200", w.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(&mock.MockFavoriteRepository{}))
		w := favReq(newFavoriteEngine(h, ""), "GET", "/favorites", nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("List status = %d, want 401", w.Code)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockFavoriteRepository{
			ListByUserFn: func(_ context.Context, _ string) ([]entity.Track, error) { return nil, errStreamTest },
		}
		h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(repo))
		w := favReq(newFavoriteEngine(h, "u1"), "GET", "/favorites", nil)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("List status = %d, want 500", w.Code)
		}
	})
}

func TestFavoriteHandler_Add(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       map[string]interface{}
		repoSetup  func(*mock.MockFavoriteRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: "u1",
			body:   map[string]interface{}{"track_id": "t1"},
			repoSetup: func(r *mock.MockFavoriteRepository) {
				r.AddFn = func(_ context.Context, _, _ string) error { return nil }
			},
			wantStatus: http.StatusCreated,
		},
		{name: "missing track_id", userID: "u1", body: map[string]interface{}{}, wantStatus: http.StatusBadRequest},
		{name: "unauthenticated", userID: "", body: map[string]interface{}{"track_id": "t1"}, wantStatus: http.StatusUnauthorized},
		{
			name:   "repo error",
			userID: "u1",
			body:   map[string]interface{}{"track_id": "t1"},
			repoSetup: func(r *mock.MockFavoriteRepository) {
				r.AddFn = func(_ context.Context, _, _ string) error { return errStreamTest }
			},
			wantStatus: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockFavoriteRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(repo))
			w := favReq(newFavoriteEngine(h, tt.userID), "POST", "/favorites", tt.body)
			if w.Code != tt.wantStatus {
				t.Fatalf("Add status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestFavoriteHandler_Remove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockFavoriteRepository{
			RemoveFn: func(_ context.Context, _, _ string) error { return nil },
		}
		h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(repo))
		w := favReq(newFavoriteEngine(h, "u1"), "DELETE", "/favorites/t1", nil)
		if w.Code != http.StatusNoContent {
			t.Fatalf("Remove status = %d, want 204", w.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(&mock.MockFavoriteRepository{}))
		w := favReq(newFavoriteEngine(h, ""), "DELETE", "/favorites/t1", nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Remove status = %d, want 401", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mock.MockFavoriteRepository{
			RemoveFn: func(_ context.Context, _, _ string) error { return errStreamTest },
		}
		h := handler.NewFavoriteHandler(usecase.NewFavoriteUseCase(repo))
		w := favReq(newFavoriteEngine(h, "u1"), "DELETE", "/favorites/t1", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("Remove status = %d, want 404", w.Code)
		}
	})
}
