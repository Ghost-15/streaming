package supabase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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

const playlistColumns = `id, owner_id, title, is_queue, track_count, created_at`

// FindByID returns the playlist with its tracks, or nil if not found.
func (r *supabasePlaylistRepo) FindByID(ctx context.Context, id string) (*entity.Playlist, error) {
	const q = `SELECT ` + playlistColumns + ` FROM playlists WHERE id = $1`
	p := &entity.Playlist{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.OwnerID, &p.Title, &p.IsQueue, &p.TrackCount, &p.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("playlist_repo: find by id: %w", err)
	}

	tracks, err := r.listTracks(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Tracks = tracks
	return p, nil
}

func (r *supabasePlaylistRepo) listTracks(ctx context.Context, playlistID string) ([]entity.Track, error) {
	const q = `
		SELECT t.id, t.title, t.artist, t.duration, t.file_url, t.uploaded_by, t.created_at,
		       pt.position, pt.added_at
		FROM playlist_tracks pt
		JOIN tracks t ON t.id = pt.track_id
		WHERE pt.playlist_id = $1
		ORDER BY pt.position ASC`
	rows, err := r.db.Query(ctx, q, playlistID)
	if err != nil {
		return nil, fmt.Errorf("playlist_repo: list tracks: %w", err)
	}
	defer rows.Close()

	tracks := []entity.Track{}
	for rows.Next() {
		var t entity.Track
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Artist, &t.Duration, &t.FileURL, &t.UploadedBy, &t.CreatedAt,
			&t.Position, &t.AddedAt,
		); err != nil {
			return nil, fmt.Errorf("playlist_repo: scan track: %w", err)
		}
		t.PlaylistID = playlistID
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}

func (r *supabasePlaylistRepo) ListByOwner(ctx context.Context, ownerID string) ([]entity.Playlist, error) {
	const q = `SELECT ` + playlistColumns + ` FROM playlists WHERE owner_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, ownerID)
	if err != nil {
		return nil, fmt.Errorf("playlist_repo: list by owner: %w", err)
	}
	defer rows.Close()

	playlists := []entity.Playlist{}
	for rows.Next() {
		var p entity.Playlist
		if err := rows.Scan(
			&p.ID, &p.OwnerID, &p.Title, &p.IsQueue, &p.TrackCount, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("playlist_repo: scan playlist: %w", err)
		}
		playlists = append(playlists, p)
	}
	return playlists, rows.Err()
}

func (r *supabasePlaylistRepo) Create(ctx context.Context, playlist *entity.Playlist) error {
	const q = `
		INSERT INTO playlists (owner_id, title, is_queue)
		VALUES ($1, $2, $3)
		RETURNING id, track_count, created_at`
	err := r.db.QueryRow(ctx, q, playlist.OwnerID, playlist.Title, playlist.IsQueue).
		Scan(&playlist.ID, &playlist.TrackCount, &playlist.CreatedAt)
	if err != nil {
		return fmt.Errorf("playlist_repo: create: %w", err)
	}
	return nil
}

func (r *supabasePlaylistRepo) Update(ctx context.Context, playlist *entity.Playlist) error {
	const q = `UPDATE playlists SET title = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, q, playlist.Title, playlist.ID)
	if err != nil {
		return fmt.Errorf("playlist_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("playlist_repo: playlist %s not found", playlist.ID)
	}
	return nil
}

func (r *supabasePlaylistRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM playlists WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("playlist_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("playlist_repo: playlist %s not found", id)
	}
	return nil
}

// AddTrack appends a track at the next position (max(position)+1).
// The track_count column is maintained by the DB trigger (migration 005).
func (r *supabasePlaylistRepo) AddTrack(ctx context.Context, track *entity.Track) error {
	const q = `
		INSERT INTO playlist_tracks (playlist_id, track_id, position)
		VALUES (
			$1, $2,
			COALESCE((SELECT MAX(position) + 1 FROM playlist_tracks WHERE playlist_id = $1), 0)
		)`
	_, err := r.db.Exec(ctx, q, track.PlaylistID, track.ID)
	if err != nil {
		return fmt.Errorf("playlist_repo: add track: %w", err)
	}
	return nil
}

func (r *supabasePlaylistRepo) RemoveTrack(ctx context.Context, playlistID, trackID string) error {
	const q = `DELETE FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`
	tag, err := r.db.Exec(ctx, q, playlistID, trackID)
	if err != nil {
		return fmt.Errorf("playlist_repo: remove track: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("playlist_repo: track %s not in playlist %s", trackID, playlistID)
	}
	return nil
}
