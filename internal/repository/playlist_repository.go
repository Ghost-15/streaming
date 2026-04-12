package repository

import (
	"context"

	"github.com/Ghost-15/streaming/internal/entity"
)

// PlaylistRepository defines the persistence contract for playlists.
// Implemented in internal/infrastructure/supabase/playlist_repo.go
type PlaylistRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Playlist, error)
	ListByOwner(ctx context.Context, ownerID string) ([]entity.Playlist, error)
	Create(ctx context.Context, playlist *entity.Playlist) error
	Update(ctx context.Context, playlist *entity.Playlist) error
	Delete(ctx context.Context, id string) error
	AddTrack(ctx context.Context, track *entity.Track) error
	RemoveTrack(ctx context.Context, playlistID, trackID string) error
}
