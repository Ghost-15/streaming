package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

func queuePlaylist() *entity.Playlist {
	return &entity.Playlist{
		ID:      "pl-1",
		OwnerID: "owner-1",
		IsQueue: true,
		Tracks: []entity.Track{
			{ID: "t1", Position: 0},
			{ID: "t2", Position: 1},
		},
	}
}

func TestPlaylistUseCase_ReorderTracks(t *testing.T) {
	tests := []struct {
		name      string
		ownerID   string
		order     []string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
	}{
		{
			name:    "success",
			ownerID: "owner-1",
			order:   []string{"t2", "t1"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) { return queuePlaylist(), nil }
				r.ReorderTracksFn = func(_ context.Context, _ string, _ []string) error { return nil }
			},
		},
		{
			name:    "empty order",
			ownerID: "owner-1",
			order:   []string{},
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "wrong length",
			ownerID: "owner-1",
			order:   []string{"t1"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) { return queuePlaylist(), nil }
			},
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "unknown track",
			ownerID: "owner-1",
			order:   []string{"t1", "tX"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) { return queuePlaylist(), nil }
			},
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "forbidden",
			ownerID: "intruder",
			order:   []string{"t2", "t1"},
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) { return queuePlaylist(), nil }
			},
			wantErr: usecase.ErrPlaylistForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)
			err := uc.ReorderTracks(context.Background(), "pl-1", tt.ownerID, tt.order)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("ReorderTracks() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaylistUseCase_NextTrack(t *testing.T) {
	t.Run("success pops head", func(t *testing.T) {
		repo := &mock.MockPlaylistRepository{
			FindByIDFn:    func(_ context.Context, _ string) (*entity.Playlist, error) { return queuePlaylist(), nil },
			RemoveTrackFn: func(_ context.Context, _, _ string) error { return nil },
		}
		uc := usecase.NewPlaylistUseCase(repo)
		track, err := uc.NextTrack(context.Background(), "pl-1", "owner-1")
		if err != nil {
			t.Fatalf("NextTrack() err = %v", err)
		}
		if track.ID != "t1" {
			t.Errorf("NextTrack() head = %q, want t1", track.ID)
		}
	})

	t.Run("empty queue", func(t *testing.T) {
		repo := &mock.MockPlaylistRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Playlist, error) {
				return &entity.Playlist{ID: "pl-1", OwnerID: "owner-1", Tracks: []entity.Track{}}, nil
			},
		}
		uc := usecase.NewPlaylistUseCase(repo)
		if _, err := uc.NextTrack(context.Background(), "pl-1", "owner-1"); !errors.Is(err, usecase.ErrPlaylistEmpty) {
			t.Fatalf("NextTrack() err = %v, want ErrPlaylistEmpty", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		repo := &mock.MockPlaylistRepository{
			FindByIDFn: func(_ context.Context, _ string) (*entity.Playlist, error) { return nil, nil },
		}
		uc := usecase.NewPlaylistUseCase(repo)
		if _, err := uc.NextTrack(context.Background(), "pl-1", "owner-1"); !errors.Is(err, usecase.ErrPlaylistNotFound) {
			t.Fatalf("NextTrack() err = %v, want ErrPlaylistNotFound", err)
		}
	})
}
