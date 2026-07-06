package supabase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

type supabaseFavoriteRepo struct {
	db *pgxpool.Pool
}

func NewFavoriteRepo(db *pgxpool.Pool) repository.FavoriteRepository {
	return &supabaseFavoriteRepo{db: db}
}

func (r *supabaseFavoriteRepo) Add(ctx context.Context, userID, trackID string) error {
	const q = `
		INSERT INTO favorites (user_id, track_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, track_id) DO NOTHING`
	if _, err := r.db.Exec(ctx, q, userID, trackID); err != nil {
		return fmt.Errorf("favorite_repo: add: %w", err)
	}
	return nil
}

func (r *supabaseFavoriteRepo) Remove(ctx context.Context, userID, trackID string) error {
	const q = `DELETE FROM favorites WHERE user_id = $1 AND track_id = $2`
	tag, err := r.db.Exec(ctx, q, userID, trackID)
	if err != nil {
		return fmt.Errorf("favorite_repo: remove: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("favorite_repo: favorite not found")
	}
	return nil
}

func (r *supabaseFavoriteRepo) ListByUser(ctx context.Context, userID string) ([]entity.Track, error) {
	const q = `
		SELECT t.id, t.title, t.artist, t.duration, t.file_url, t.uploaded_by, t.created_at
		FROM favorites f
		JOIN tracks t ON t.id = f.track_id
		WHERE f.user_id = $1
		ORDER BY f.created_at DESC`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite_repo: list: %w", err)
	}
	defer rows.Close()

	tracks := []entity.Track{}
	for rows.Next() {
		var t entity.Track
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Artist, &t.Duration, &t.FileURL, &t.UploadedBy, &t.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("favorite_repo: scan track: %w", err)
		}
		tracks = append(tracks, t)
	}
	return tracks, rows.Err()
}
