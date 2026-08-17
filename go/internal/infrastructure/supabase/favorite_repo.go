package supabase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

type supabaseFavoriteRepo struct {
	db pgxDB
}

func NewFavoriteRepo(db *pgxpool.Pool) repository.FavoriteRepository {
	return &supabaseFavoriteRepo{db: poolOrNil(db)}
}

func (r *supabaseFavoriteRepo) Add(ctx context.Context, userID, trackID string) error {
	var err error
	ctx, span := startRepoSpan(ctx, "favorites", "FavoriteRepository", "Add", "favorites", "INSERT")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		INSERT INTO favorites (user_id, track_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, track_id) DO NOTHING`
	if _, err = r.db.Exec(ctx, q, userID, trackID); err != nil {
		return fmt.Errorf("favorite_repo: add: %w", err)
	}
	return nil
}

func (r *supabaseFavoriteRepo) Remove(ctx context.Context, userID, trackID string) error {
	var err error
	ctx, span := startRepoSpan(ctx, "favorites", "FavoriteRepository", "Remove", "favorites", "DELETE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `DELETE FROM favorites WHERE user_id = $1 AND track_id = $2`
	tag, err := r.db.Exec(ctx, q, userID, trackID)
	if err != nil {
		return fmt.Errorf("favorite_repo: remove: %w", err)
	}
	span.SetAttributes(attribute.Int64("db.result.rows_affected", tag.RowsAffected()))
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("favorite_repo: favorite not found")
	}
	return nil
}

func (r *supabaseFavoriteRepo) ListByUser(ctx context.Context, userID string) ([]entity.Track, error) {
	var err error
	ctx, span := startRepoSpan(ctx, "favorites", "FavoriteRepository", "ListByUser", "favorites", "SELECT",
		attribute.String("lookup.kind", "user_id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `
		SELECT s.id, s.title,
		       COALESCE(NULLIF(TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')), ''), '') AS artist,
		       0 AS duration, '' AS file_url, s.broadcaster_id AS uploaded_by, f.created_at
		FROM favorites f
		JOIN streams s ON s.id = f.track_id
		LEFT JOIN users u ON u.id = s.broadcaster_id
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
	span.SetAttributes(attribute.Int("db.result.row_count", len(tracks)))
	if rowsErr := rows.Err(); rowsErr != nil {
		err = rowsErr
		return nil, err
	}
	return tracks, nil
}
