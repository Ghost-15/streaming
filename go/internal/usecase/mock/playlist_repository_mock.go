package mock

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// Compile-time check: MockPlaylistRepository implements repository.PlaylistRepository.
var _ repository.PlaylistRepository = (*MockPlaylistRepository)(nil)

// MockPlaylistRepository is a hand-rolled mock for usecase tests.
// Set the function fields to control behavior per test case.
type MockPlaylistRepository struct {
	FindByIDFn    func(ctx context.Context, id string) (*entity.Playlist, error)
	ListByOwnerFn func(ctx context.Context, ownerID string) ([]entity.Playlist, error)
	CreateFn      func(ctx context.Context, playlist *entity.Playlist) error
	UpdateFn      func(ctx context.Context, playlist *entity.Playlist) error
	DeleteFn      func(ctx context.Context, id string) error
	AddTrackFn    func(ctx context.Context, track *entity.Track) error
	RemoveTrackFn func(ctx context.Context, playlistID, trackID string) error
}

func (m *MockPlaylistRepository) FindByID(ctx context.Context, id string) (*entity.Playlist, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *MockPlaylistRepository) ListByOwner(ctx context.Context, ownerID string) ([]entity.Playlist, error) {
	return m.ListByOwnerFn(ctx, ownerID)
}

func (m *MockPlaylistRepository) Create(ctx context.Context, playlist *entity.Playlist) error {
	return m.CreateFn(ctx, playlist)
}

func (m *MockPlaylistRepository) Update(ctx context.Context, playlist *entity.Playlist) error {
	return m.UpdateFn(ctx, playlist)
}

func (m *MockPlaylistRepository) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *MockPlaylistRepository) AddTrack(ctx context.Context, track *entity.Track) error {
	return m.AddTrackFn(ctx, track)
}

func (m *MockPlaylistRepository) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	return m.RemoveTrackFn(ctx, playlistID, trackID)
}
