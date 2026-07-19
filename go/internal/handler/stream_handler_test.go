package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/handler"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

var errStreamTest = errors.New("stream test error")

func newStreamEngine(h *handler.StreamHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if userID != "" {
		r.Use(func(c *gin.Context) {
			c.Set("claims", &entity.JWTClaims{UserID: userID, Role: entity.RoleDiffuseur})
			c.Next()
		})
	}
	r.GET("/streams", h.ListActive)
	r.POST("/streams", h.Start)
	r.PUT("/streams/:id/stop", h.Stop)
	r.POST("/streams/:id/listen", h.Listen)
	r.POST("/streams/:id/leave", h.Leave)
	return r
}

func newStreamHandler(streamRepo *mock.MockStreamRepository) *handler.StreamHandler {
	return handler.NewStreamHandler(usecase.NewStreamUseCase(streamRepo, nil))
}

func newStreamHandlerWithHistory(streamRepo *mock.MockStreamRepository, historyRepo *mock.MockListenHistoryRepository) *handler.StreamHandler {
	return handler.NewStreamHandler(usecase.NewStreamUseCase(streamRepo, historyRepo))
}

func streamReq(engine *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var r *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	} else {
		r = bytes.NewReader(nil)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, r)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func TestStreamHandler_ListActive(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			ListActiveFn: func(_ context.Context) ([]entity.Stream, error) {
				return []entity.Stream{{ID: "s1", Status: entity.StreamStatusLive}}, nil
			},
		}
		w := streamReq(newStreamEngine(newStreamHandler(repo), ""), "GET", "/streams", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("ListActive status = %d, want 200", w.Code)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			ListActiveFn: func(_ context.Context) ([]entity.Stream, error) { return nil, errStreamTest },
		}
		w := streamReq(newStreamEngine(newStreamHandler(repo), ""), "GET", "/streams", nil)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("ListActive status = %d, want 500", w.Code)
		}
	})
}

func TestStreamHandler_Start(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       map[string]interface{}
		repoSetup  func(*mock.MockStreamRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: "bc-1",
			body:   map[string]interface{}{"title": "My Live Show"},
			repoSetup: func(r *mock.MockStreamRepository) {
				r.CreateFn = func(_ context.Context, s *entity.Stream) error { s.ID = "s1"; return nil }
			},
			wantStatus: http.StatusCreated,
		},
		{name: "missing claims", userID: "", body: map[string]interface{}{"title": "My Live Show"}, wantStatus: http.StatusUnauthorized},
		{name: "invalid title", userID: "bc-1", body: map[string]interface{}{"title": "ab"}, wantStatus: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			w := streamReq(newStreamEngine(newStreamHandler(repo), tt.userID), "POST", "/streams", tt.body)
			if w.Code != tt.wantStatus {
				t.Fatalf("Start status = %d, want %d (body=%s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestStreamHandler_Stop(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repoSetup  func(*mock.MockStreamRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: "bc-1",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) {
					return &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusLive}, nil
				}
				r.UpdateStatusFn = func(_ context.Context, _ string, _ entity.StreamStatus) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{name: "missing claims", userID: "", wantStatus: http.StatusUnauthorized},
		{
			name:   "not found",
			userID: "bc-1",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) { return nil, nil }
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "forbidden",
			userID: "other",
			repoSetup: func(r *mock.MockStreamRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Stream, error) {
					return &entity.Stream{ID: "s1", BroadcasterID: "bc-1", Status: entity.StreamStatusLive}, nil
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStreamRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			w := streamReq(newStreamEngine(newStreamHandler(repo), tt.userID), "PUT", "/streams/s1/stop", nil)
			if w.Code != tt.wantStatus {
				t.Fatalf("Stop status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestStreamHandler_Listen(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) {
				return &entity.Stream{ID: "s1", Status: entity.StreamStatusLive}, nil
			},
			IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return nil },
		}
		w := streamReq(newStreamEngine(newStreamHandler(repo), "u1"), "POST", "/streams/s1/listen", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("Listen status = %d, want 200", w.Code)
		}
	})

	t.Run("missing claims", func(t *testing.T) {
		w := streamReq(newStreamEngine(newStreamHandler(&mock.MockStreamRepository{}), ""), "POST", "/streams/s1/listen", nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Listen status = %d, want 401", w.Code)
		}
	})

	t.Run("stream not found", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Stream, error) { return nil, nil },
		}
		w := streamReq(newStreamEngine(newStreamHandler(repo), "u1"), "POST", "/streams/s1/listen", nil)
		if w.Code != http.StatusNotFound {
			t.Fatalf("Listen status = %d, want 404", w.Code)
		}
	})
}

func TestStreamHandler_Leave(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			IncrementListenersFn: func(_ context.Context, id string, delta int) error {
				if id != "s1" || delta != -1 {
					t.Fatalf("IncrementListeners(%q, %d), want s1, -1", id, delta)
				}
				return nil
			},
		}
		history := &mock.MockListenHistoryRepository{
			RecordFn: func(_ context.Context, entry *entity.ListenHistory) error {
				if entry.Event != entity.ListenEventLeave {
					t.Fatalf("history event = %q, want leave", entry.Event)
				}
				return nil
			},
		}
		w := streamReq(newStreamEngine(newStreamHandlerWithHistory(repo, history), "u1"), "POST", "/streams/s1/leave", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("Leave status = %d, want 200", w.Code)
		}
	})

	t.Run("missing claims", func(t *testing.T) {
		w := streamReq(newStreamEngine(newStreamHandler(&mock.MockStreamRepository{}), ""), "POST", "/streams/s1/leave", nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Leave status = %d, want 401", w.Code)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockStreamRepository{
			IncrementListenersFn: func(_ context.Context, _ string, _ int) error { return errStreamTest },
		}
		w := streamReq(newStreamEngine(newStreamHandler(repo), "u1"), "POST", "/streams/s1/leave", nil)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("Leave status = %d, want 500", w.Code)
		}
	})
}
