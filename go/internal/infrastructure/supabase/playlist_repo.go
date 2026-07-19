package supabase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// supabasePlaylistRepo implements repository.PlaylistRepository.
type supabasePlaylistRepo struct {
	db pgxDB
}

// NewPlaylistRepo returns a PlaylistRepository backed by Supabase.
func NewPlaylistRepo(db *pgxpool.Pool) repository.PlaylistRepository {
	return &supabasePlaylistRepo{db: poolOrNil(db)}
}

const playlistColumns = `id, owner_id, title, is_queue, track_count, created_at`

func (r *supabasePlaylistRepo) FindByID(ctx context.Context, id string) (playlist *entity.Playlist, err error) {
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "FindByID", "playlists", "SELECT",
		attribute.String("lookup.kind", "id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + playlistColumns + ` FROM playlists WHERE id = $1`
	p := &entity.Playlist{}
	scanErr := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.OwnerID, &p.Title, &p.IsQueue, &p.TrackCount, &p.CreatedAt,
	)
	if errors.Is(scanErr, pgx.ErrNoRows) {
		return nil, nil
	}
	if scanErr != nil {
		err = fmt.Errorf("playlist_repo: find by id: %w", scanErr)
		return nil, err
	}

	tracks, tErr := r.listTracks(ctx, id)
	if tErr != nil {
		err = tErr
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

func (r *supabasePlaylistRepo) ListByOwner(ctx context.Context, ownerID string) (playlists []entity.Playlist, err error) {
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "ListByOwner", "playlists", "SELECT",
		attribute.String("lookup.kind", "owner_id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + playlistColumns + ` FROM playlists WHERE owner_id = $1 ORDER BY created_at DESC`
	rows, queryErr := r.db.Query(ctx, q, ownerID)
	if queryErr != nil {
		err = fmt.Errorf("playlist_repo: list by owner: %w", queryErr)
		return nil, err
	}
	defer rows.Close()

	playlists = []entity.Playlist{}
	for rows.Next() {
		var p entity.Playlist
		if scanErr := rows.Scan(
			&p.ID, &p.OwnerID, &p.Title, &p.IsQueue, &p.TrackCount, &p.CreatedAt,
		); scanErr != nil {
			err = fmt.Errorf("playlist_repo: scan playlist: %w", scanErr)
			return nil, err
		}
		playlists = append(playlists, p)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		err = rowsErr
		return nil, err
	}

	span.SetAttributes(attribute.Int("db.result.row_count", len(playlists)))
	return playlists, nil
}

func (r *supabasePlaylistRepo) Create(ctx context.Context, playlist *entity.Playlist) (err error) {
	attrs := playlistAttrs(playlist)
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "Create", "playlists", "INSERT", attrs...)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		INSERT INTO playlists (owner_id, title, is_queue)
		VALUES ($1, $2, $3)
		RETURNING id, track_count, created_at`
	err = r.db.QueryRow(ctx, q, playlist.OwnerID, playlist.Title, playlist.IsQueue).
		Scan(&playlist.ID, &playlist.TrackCount, &playlist.CreatedAt)
	if err != nil {
		return fmt.Errorf("playlist_repo: create: %w", err)
	}
	return nil
}

func (r *supabasePlaylistRepo) Update(ctx context.Context, playlist *entity.Playlist) (err error) {
	attrs := playlistAttrs(playlist)
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "Update", "playlists", "UPDATE", attrs...)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `UPDATE playlists SET title = $1 WHERE id = $2`
	tag, execErr := r.db.Exec(ctx, q, playlist.Title, playlist.ID)
	if execErr != nil {
		return fmt.Errorf("playlist_repo: update: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("playlist_repo: playlist %s not found", playlist.ID)
	}
	return nil
}

func (r *supabasePlaylistRepo) Delete(ctx context.Context, id string) (err error) {
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "Delete", "playlists", "DELETE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `DELETE FROM playlists WHERE id = $1`
	tag, execErr := r.db.Exec(ctx, q, id)
	if execErr != nil {
		return fmt.Errorf("playlist_repo: delete: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("playlist_repo: playlist %s not found", id)
	}
	return nil
}

func (r *supabasePlaylistRepo) AddTrack(ctx context.Context, track *entity.Track) (err error) {
	attrs := []attribute.KeyValue{}
	if track != nil {
		attrs = append(attrs,
			attribute.Int("track.duration_seconds", track.Duration),
			attribute.Int("track.position", track.Position),
		)
	}
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "AddTrack", "playlist_tracks", "INSERT", attrs...)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		INSERT INTO playlist_tracks (playlist_id, track_id, position)
		VALUES (
			$1, $2,
			COALESCE((SELECT MAX(position) + 1 FROM playlist_tracks WHERE playlist_id = $1), 0)
		)`
	if _, execErr := r.db.Exec(ctx, q, track.PlaylistID, track.ID); execErr != nil {
		return fmt.Errorf("playlist_repo: add track: %w", execErr)
	}
	return nil
}

func (r *supabasePlaylistRepo) RemoveTrack(ctx context.Context, playlistID, trackID string) (err error) {
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "RemoveTrack", "playlist_tracks", "DELETE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `DELETE FROM playlist_tracks WHERE playlist_id = $1 AND track_id = $2`
	tag, execErr := r.db.Exec(ctx, q, playlistID, trackID)
	if execErr != nil {
		return fmt.Errorf("playlist_repo: remove track: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("playlist_repo: track %s not in playlist %s", trackID, playlistID)
	}
	return nil
}

func (r *supabasePlaylistRepo) ReorderTracks(ctx context.Context, playlistID string, orderedTrackIDs []string) (err error) {
	ctx, span := startRepoSpan(ctx, "playlists", "PlaylistRepository", "ReorderTracks", "playlist_tracks", "UPDATE",
		attribute.Int("playlist.track_count", len(orderedTrackIDs)),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	tx, txErr := r.db.Begin(ctx)
	if txErr != nil {
		err = fmt.Errorf("playlist_repo: reorder begin: %w", txErr)
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, execErr := tx.Exec(ctx,
		`UPDATE playlist_tracks SET position = position + 1000000 WHERE playlist_id = $1`,
		playlistID,
	); execErr != nil {
		err = fmt.Errorf("playlist_repo: reorder offset: %w", execErr)
		return err
	}

	for i, trackID := range orderedTrackIDs {
		tag, execErr := tx.Exec(ctx,
			`UPDATE playlist_tracks SET position = $3 WHERE playlist_id = $1 AND track_id = $2`,
			playlistID, trackID, i,
		)
		if execErr != nil {
			err = fmt.Errorf("playlist_repo: reorder set: %w", execErr)
			return err
		}
		if tag.RowsAffected() == 0 {
			err = fmt.Errorf("playlist_repo: track %s not in playlist %s", trackID, playlistID)
			return err
		}
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = fmt.Errorf("playlist_repo: reorder commit: %w", commitErr)
		return err
	}
	return nil
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
