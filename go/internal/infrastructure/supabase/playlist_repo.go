package supabase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

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

func (r *supabasePlaylistRepo) FindByID(ctx context.Context, id string) (playlist *entity.Playlist, err error) {
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "FindByID", "playlists", "SELECT",
		attribute.String("lookup.kind", "id"),
	)
	defer finishRepoSpan(span, &err)

	return nil, errors.New("not implemented")
}

func (r *supabasePlaylistRepo) ListByOwner(ctx context.Context, ownerID string) (playlists []entity.Playlist, err error) {
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "ListByOwner", "playlists", "SELECT",
		attribute.String("lookup.kind", "owner_id"),
	)
	defer finishRepoSpan(span, &err)

	return nil, errors.New("not implemented")
}

func (r *supabasePlaylistRepo) Create(ctx context.Context, playlist *entity.Playlist) (err error) {
	attrs := playlistAttrs(playlist)
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "Create", "playlists", "INSERT", attrs...)
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) Update(ctx context.Context, playlist *entity.Playlist) (err error) {
	attrs := playlistAttrs(playlist)
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "Update", "playlists", "UPDATE", attrs...)
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) Delete(ctx context.Context, id string) (err error) {
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "Delete", "playlists", "DELETE")
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) AddTrack(ctx context.Context, track *entity.Track) (err error) {
	attrs := []attribute.KeyValue{}
	if track != nil {
		attrs = append(attrs,
			attribute.Int("track.duration_seconds", track.Duration),
			attribute.Int("track.position", track.Position),
		)
	}
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "AddTrack", "playlist_tracks", "INSERT", attrs...)
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func (r *supabasePlaylistRepo) RemoveTrack(ctx context.Context, playlistID, trackID string) (err error) {
	_, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "RemoveTrack", "playlist_tracks", "DELETE")
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func playlistAttrs(playlist *entity.Playlist) []attribute.KeyValue {
	if playlist == nil {
		return nil
	}
	return []attribute.KeyValue{
		attribute.Bool("playlist.is_queue", playlist.IsQueue),
		attribute.Int("playlist.track_count", playlist.TrackCount),
	}
}
