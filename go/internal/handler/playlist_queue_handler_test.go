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

func newQueueEngine(h *handler.PlaylistHandler, userID string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	if userID != "" {
		r.Use(injectClaims(userID))
	}
	r.PUT("/playlists/:id/tracks/reorder", h.ReorderTracks)
	r.POST("/playlists/:id/next", h.NextTrack)
	return r
}

func queuePlaylistFixture(owner string) *entity.Playlist {
	return &entity.Playlist{
		ID:      "pl-1",
		OwnerID: owner,
		IsQueue: true,
		Tracks:  []entity.Track{{ID: "t1", Position: 0}, {ID: "t2", Position: 1}},
	}
}

func TestPlaylistHandler_ReorderTracks(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		body       map[string]interface{}
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: "owner-1",
			body:   map[string]interface{}{"track_ids": []string{"t2", "t1"}},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return queuePlaylistFixture("owner-1"), nil
				}
				r.ReorderTracksFn = func(_ context.Context, _ string, _ []string) error { return nil }
			},
			wantStatus: http.StatusNoContent,
		},
		{name: "missing claims", userID: "", body: map[string]interface{}{"track_ids": []string{"t1"}}, wantStatus: http.StatusUnauthorized},
		{name: "invalid payload", userID: "owner-1", body: map[string]interface{}{"track_ids": []string{}}, wantStatus: http.StatusBadRequest},
		{
			name:   "forbidden",
			userID: "intruder",
			body:   map[string]interface{}{"track_ids": []string{"t2", "t1"}},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return queuePlaylistFixture("owner-1"), nil
				}
			},
			wantStatus: http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			b, _ := json.Marshal(tt.body)
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/playlists/pl-1/tracks/reorder", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			newQueueEngine(h, tt.userID).ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("Reorder status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestPlaylistHandler_NextTrack(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		repoSetup  func(*mock.MockPlaylistRepository)
		wantStatus int
	}{
		{
			name:   "success",
			userID: "owner-1",
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return queuePlaylistFixture("owner-1"), nil
				}
				r.RemoveTrackFn = func(_ context.Context, _, _ string) error { return nil }
			},
			wantStatus: http.StatusOK,
		},
		{name: "missing claims", userID: "", wantStatus: http.StatusUnauthorized},
		{
			name:   "empty queue",
			userID: "owner-1",
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: "pl-1", OwnerID: "owner-1", Tracks: []entity.Track{}}, nil
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			h := handler.NewPlaylistHandler(usecase.NewPlaylistUseCase(repo))
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/playlists/pl-1/next", nil)
			w := httptest.NewRecorder()
			newQueueEngine(h, tt.userID).ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("NextTrack status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
