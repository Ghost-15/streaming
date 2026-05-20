package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/usecase"
	"github.com/Ghost-15/streaming/internal/usecase/mock"
)

const (
	testOwnerID    = "owner-123"
	testOtherOwner = "intruder-456"
	testPlaylistID = "playlist-uuid-1"
	testTrackID    = "track-uuid-1"
)

// ─────────────────────────────────────────────────────────────
// Create
// ─────────────────────────────────────────────────────────────

func TestPlaylistUseCase_Create(t *testing.T) {
	tests := []struct {
		name      string
		ownerID   string
		title     string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
	}{
		{
			name:    "success",
			ownerID: testOwnerID,
			title:   "My Playlist",
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.CreateFn = func(_ context.Context, _ *entity.Playlist) error {
					return nil
				}
			},
		},
		{
			name:    "empty title",
			ownerID: testOwnerID,
			title:   "   ",
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "empty owner",
			ownerID: "",
			title:   "Whatever",
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "repository error",
			ownerID: testOwnerID,
			title:   "Boom",
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.CreateFn = func(_ context.Context, _ *entity.Playlist) error {
					return errors.New("db down")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			pl, err := uc.Create(context.Background(), tt.ownerID, tt.title)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Create() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil && tt.repoSetup == nil {
				t.Fatalf("Create() unexpected err = %v", err)
			}
			if err == nil {
				if pl == nil {
					t.Fatal("Create() returned nil playlist on success")
				}
				if pl.ID == "" {
					t.Error("Create() did not generate an ID")
				}
				if pl.OwnerID != tt.ownerID {
					t.Errorf("Create() ownerID = %q, want %q", pl.OwnerID, tt.ownerID)
				}
				if pl.Title != "My Playlist" {
					t.Errorf("Create() title = %q, want %q", pl.Title, "My Playlist")
				}
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// List
// ─────────────────────────────────────────────────────────────

func TestPlaylistUseCase_List(t *testing.T) {
	tests := []struct {
		name      string
		ownerID   string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
		wantLen   int
	}{
		{
			name:    "success — empty",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.ListByOwnerFn = func(_ context.Context, _ string) ([]entity.Playlist, error) {
					return []entity.Playlist{}, nil
				}
			},
		},
		{
			name:    "success — two playlists",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.ListByOwnerFn = func(_ context.Context, _ string) ([]entity.Playlist, error) {
					return []entity.Playlist{{ID: "a"}, {ID: "b"}}, nil
				}
			},
			wantLen: 2,
		},
		{
			name:    "empty owner",
			ownerID: "",
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "repository error",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.ListByOwnerFn = func(_ context.Context, _ string) ([]entity.Playlist, error) {
					return nil, errors.New("db timeout")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			out, err := uc.List(context.Background(), tt.ownerID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("List() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil && tt.repoSetup != nil && tt.name != "repository error" {
				t.Fatalf("List() unexpected err = %v", err)
			}
			if err == nil && len(out) != tt.wantLen {
				t.Errorf("List() len = %d, want %d", len(out), tt.wantLen)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// GetByID / fetchOwned coverage
// ─────────────────────────────────────────────────────────────

func TestPlaylistUseCase_GetByID(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		requesterID string
		repoSetup   func(*mock.MockPlaylistRepository)
		wantErr     error
	}{
		{
			name:        "success",
			id:          testPlaylistID,
			requesterID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
			},
		},
		{
			name:        "not found",
			id:          testPlaylistID,
			requesterID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return nil, nil
				}
			},
			wantErr: usecase.ErrPlaylistNotFound,
		},
		{
			name:        "forbidden — wrong owner",
			id:          testPlaylistID,
			requesterID: testOtherOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
			},
			wantErr: usecase.ErrPlaylistForbidden,
		},
		{
			name:        "invalid input",
			id:          "",
			requesterID: testOwnerID,
			wantErr:     usecase.ErrPlaylistInvalid,
		},
		{
			name:        "repository error",
			id:          testPlaylistID,
			requesterID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return nil, errors.New("db error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			pl, err := uc.GetByID(context.Background(), tt.id, tt.requesterID)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("GetByID() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err == nil && pl == nil {
				t.Fatal("GetByID() returned nil playlist on success")
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────────────────────

func TestPlaylistUseCase_Update(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		ownerID   string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
	}{
		{
			name:    "success",
			title:   "Renamed",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID, Title: "Old"}, nil
				}
				r.UpdateFn = func(_ context.Context, _ *entity.Playlist) error { return nil }
			},
		},
		{
			name:    "empty title",
			title:   "   ",
			ownerID: testOwnerID,
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "forbidden",
			title:   "Whatever",
			ownerID: testOtherOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
			},
			wantErr: usecase.ErrPlaylistForbidden,
		},
		{
			name:    "update repo error",
			title:   "Renamed",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.UpdateFn = func(_ context.Context, _ *entity.Playlist) error {
					return errors.New("db down")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			pl, err := uc.Update(context.Background(), testPlaylistID, tt.ownerID, tt.title)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Update() err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err == nil && pl.Title != "Renamed" {
				t.Errorf("Update() title = %q, want %q", pl.Title, "Renamed")
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// Delete
// ─────────────────────────────────────────────────────────────

func TestPlaylistUseCase_Delete(t *testing.T) {
	tests := []struct {
		name      string
		ownerID   string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
	}{
		{
			name:    "success",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.DeleteFn = func(_ context.Context, _ string) error { return nil }
			},
		},
		{
			name:    "not found",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return nil, nil
				}
			},
			wantErr: usecase.ErrPlaylistNotFound,
		},
		{
			name:    "delete repo error",
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.DeleteFn = func(_ context.Context, _ string) error {
					return errors.New("constraint violation")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			err := uc.Delete(context.Background(), testPlaylistID, tt.ownerID)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("Delete() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────
// AddTrack / RemoveTrack
// ─────────────────────────────────────────────────────────────

func TestPlaylistUseCase_AddTrack(t *testing.T) {
	tests := []struct {
		name      string
		trackID   string
		ownerID   string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
	}{
		{
			name:    "success",
			trackID: testTrackID,
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.AddTrackFn = func(_ context.Context, _ *entity.Track) error { return nil }
			},
		},
		{
			name:    "empty trackID",
			trackID: "",
			ownerID: testOwnerID,
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "forbidden",
			trackID: testTrackID,
			ownerID: testOtherOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
			},
			wantErr: usecase.ErrPlaylistForbidden,
		},
		{
			name:    "repo error",
			trackID: testTrackID,
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.AddTrackFn = func(_ context.Context, _ *entity.Track) error {
					return errors.New("duplicate")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			err := uc.AddTrack(context.Background(), testPlaylistID, tt.ownerID, tt.trackID)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("AddTrack() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlaylistUseCase_RemoveTrack(t *testing.T) {
	tests := []struct {
		name      string
		trackID   string
		ownerID   string
		repoSetup func(*mock.MockPlaylistRepository)
		wantErr   error
	}{
		{
			name:    "success",
			trackID: testTrackID,
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.RemoveTrackFn = func(_ context.Context, _, _ string) error { return nil }
			},
		},
		{
			name:    "empty trackID",
			trackID: "",
			ownerID: testOwnerID,
			wantErr: usecase.ErrPlaylistInvalid,
		},
		{
			name:    "forbidden",
			trackID: testTrackID,
			ownerID: testOtherOwner,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
			},
			wantErr: usecase.ErrPlaylistForbidden,
		},
		{
			name:    "repo error",
			trackID: testTrackID,
			ownerID: testOwnerID,
			repoSetup: func(r *mock.MockPlaylistRepository) {
				r.FindByIDFn = func(_ context.Context, _ string) (*entity.Playlist, error) {
					return &entity.Playlist{ID: testPlaylistID, OwnerID: testOwnerID}, nil
				}
				r.RemoveTrackFn = func(_ context.Context, _, _ string) error {
					return errors.New("not present")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockPlaylistRepository{}
			if tt.repoSetup != nil {
				tt.repoSetup(repo)
			}
			uc := usecase.NewPlaylistUseCase(repo)

			err := uc.RemoveTrack(context.Background(), testPlaylistID, tt.ownerID, tt.trackID)
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Fatalf("RemoveTrack() err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
