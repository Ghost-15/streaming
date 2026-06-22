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

const (
	plOwner   = "owner-uuid-A"
	plOther   = "owner-uuid-B"
	plID      = "playlist-uuid-1"
	plTrackID = "track-uuid-1"
	claimsCtx = "claims"
)

// injectClaims is a tiny middleware that simulates RBACMiddleware by setting
// the claims into the context. Used only in handler tests.
func injectClaims(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(claimsCtx, &entity.JWTClaims{
			UserID: userID,
			Role:   entity.RoleUser,
		})
		c.Next()
	}
}

// buildEngine wires the handler with a fake auth middleware injecting the given userID.
// If userID is empty, no claims are injected (simulates an unauthenticated request).
func buildEngine(h *handler.PlaylistHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if userID != "" {
		r.Use(injectClaims(userID))
	}
	r.GET("/playlists", h.List)
	r.POST("/playlists", h.Create)
	r.GET("/playlists/:id", h.GetByID)
	r.PUT("/playlists/:id", h.Update)
	r.DELETE("/playlists/:id", h.Delete)
	r.POST("/playlists/:id/tracks", h.AddTrack)
	r.DELETE("/playlists/:id/tracks/:trackID", h.RemoveTrack)
	return r
}

func doJSON(engine *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

// ─────────────────────────────────────────────────────────────
// List
// ─────────────────────────────────────────────────────────────

func TestPlaylistHandler_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mock.MockPlaylistRepository{
			ListByOwnerFn: func(_ context.Context, _ string) ([]entity.Playlist, error) {
				return []entity.Playlist{{ID: "a", OwnerID: plOwner, Title: "X"}}, nil
			},
		}
		h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
		w := doJSON(buildEngine(h, plOwner), "GET", "/playlists", nil)

		if w.Code != http.StatusOK {
			t.Fatalf("List status = %d, want 200", w.Code)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		repo := &mock.MockPlaylistRepository{}
		h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
		w := doJSON(buildEngine(h, ""), "GET", "/playlists", nil)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("List status = %d, want 401", w.Code)
		}
	})

	t.Run("repo error", func(t *testing.T) {
		repo := &mock.MockPlaylistRepository{
			ListByOwnerFn: func(_ context.Context, _ string) ([]entity.Playlist, error) {
				return nil, errors.New("db down")
			},
		}
		h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
		w := doJSON(buildEngine(h, plOwner), "GET", "/playlists", nil)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf("List status = %d, want 500", w.Code)
		}
	})
}

// ─────────────────────────────────────────────────────────────
// Create
// ─────────────────────────────────────────────────────────────

func TestPlaylistHandler_Create(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       map[string]interface{}
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: plOwner,
			body:   map[string]interface{}{"title": "Chill Vibes"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.CreateFn = func(_ context.Context, _ *entity.Playlist) error { return nil }
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing title",
			userID:     plOwner,
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "unauthenticated",
			userID:     "",
			body:       map[string]interface{}{"title": "Nope"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:   "repo error",
			userID: plOwner,
			body:   map[string]interface{}{"title": "Boom"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.CreateFn = func(_ context.Context, _ *entity.Playlist) error {
					return errors.New("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			w := doJSON(buildEngine(h, tt.userID), "POST", "/playlists", tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("Create status = %d, want %d (body=%s)", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// GetByID
// ─────────────────────────────────────────────────────────────

func TestPlaylistHandler_GetByID(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: plOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner, Title: "X"}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "not found",
			userID: plOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return nil, nil
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:   "forbidden",
			userID: plOther,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner}, nil
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unauthenticated",
			userID:     "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			w := doJSON(buildEngine(h, tt.userID), "GET", "/playlists/"+plID, nil)

			if w.Code != tt.wantStatus {
				t.Fatalf("GetByID status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────────────────────

func TestPlaylistHandler_Update(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       map[string]interface{}
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: plOwner,
			body:   map[string]interface{}{"title": "Renamed"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner, Title: "Old"}, nil
				}
				r.UpdateFn = func(_ context.Context, _ *entity.Playlist) error { return nil }
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing title",
			userID:     plOwner,
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "forbidden",
			userID: plOther,
			body:   map[string]interface{}{"title": "Hack"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner}, nil
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unauthenticated",
			userID:     "",
			body:       map[string]interface{}{"title": "X"},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			w := doJSON(buildEngine(h, tt.userID), "PUT", "/playlists/"+plID, tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("Update status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Delete
// ─────────────────────────────────────────────────────────────

func TestPlaylistHandler_Delete(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: plOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner}, nil
				}
				r.DeleteFn = func(_ context.Context, _ string) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "not found",
			userID: plOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return nil, nil
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unauthenticated",
			userID:     "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			w := doJSON(buildEngine(h, tt.userID), "DELETE", "/playlists/"+plID, nil)

			if w.Code != tt.wantStatus {
				t.Fatalf("Delete status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// AddTrack / RemoveTrack
// ─────────────────────────────────────────────────────────────

func TestPlaylistHandler_AddTrack(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       map[string]interface{}
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: plOwner,
			body:   map[string]interface{}{"track_id": plTrackID},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner}, nil
				}
				r.AddTrackFn = func(_ context.Context, _ *entity.Track) error { return nil }
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing track_id",
			userID:     plOwner,
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "forbidden",
			userID: plOther,
			body:   map[string]interface{}{"track_id": plTrackID},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner}, nil
				}
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "unauthenticated",
			userID:     "",
			body:       map[string]interface{}{"track_id": plTrackID},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			w := doJSON(buildEngine(h, tt.userID), "POST", "/playlists/"+plID+"/tracks", tt.body)

			if w.Code != tt.wantStatus {
				t.Fatalf("AddTrack status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestPlaylistHandler_RemoveTrack(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: plOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: plID, OwnerID: plOwner}, nil
				}
				r.RemoveTrackFn = func(_ context.Context, _, _ string) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:   "not found",
			userID: plOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return nil, nil
				}
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unauthenticated",
			userID:     "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			w := doJSON(buildEngine(h, tt.userID), "DELETE", "/playlists/"+plID+"/tracks/"+plTrackID, nil)

			if w.Code != tt.wantStatus {
				t.Fatalf("RemoveTrack status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
