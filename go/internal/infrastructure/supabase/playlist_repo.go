package supabase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// supabasePlaylistRepo implements repository.PlaylistRepository.
type supabasePlaylistRepo struct {
	db *pgxpool.Pool
}

// NewPlaylistRepo returns a PlaylistRepository backed by Supabase.
func NewPlaylistRepo(db *pgxpool.Pool) repository.PlaylistRepository {
	return &supabasePlaylistRepo{db: db}
}

func (r *supabasePlaylistRepo) FindByID(ctx context.Context, id string) (*entity.Playlist, error) {
	// TODO Sprint 2 — US-007
	return nil, errors.New("not implemented")
}

func (r *supabasePlaylistRepo) ListByOwner(ctx context.Context, ownerID string) ([]entity.Playlist, error) {
	// TODO Sprint 2 — US-007
	return nil, errors.New("not implemented")
}

func (r *supabasePlaylistRepo) Create(ctx context.Context, playlist *entity.Playlist) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) Update(ctx context.Context, playlist *entity.Playlist) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) Delete(ctx context.Context, id string) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) AddTrack(ctx context.Context, track *entity.Track) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	// TODO Sprint 2 — US-007
	return errors.New("not implemented")
}
