package supabase

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

type supabaseListenHistoryRepo struct {
	db *pgxpool.Pool
}

func NewListenHistoryRepo(db *pgxpool.Pool) repository.ListenHistoryRepository {
	return &supabaseListenHistoryRepo{db: db}
}

func (r *supabaseListenHistoryRepo) Record(ctx context.Context, entry *entity.ListenHistory) (err error) {
	ctx, span := startRepoSpan(ctx, "listen_history", "ListenHistoryRepository", "Record", "listen_history", "INSERT")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}
	if entry == nil {
		return fmt.Errorf("listen_history_repo: nil entry")
	}

	var trackID interface{}
	if entry.TrackID != "" {
		trackID = entry.TrackID
	}

	const q = `
		INSERT INTO listen_history (user_id, track_id, stream_id, duration_sec)
		VALUES ($1, $2, $3, $4)
		RETURNING id, listened_at`
	if err = r.db.QueryRow(ctx, q, entry.UserID, trackID, entry.StreamID, entry.DurationSec).
		Scan(&entry.ID, &entry.ListenedAt); err != nil {
		return fmt.Errorf("listen_history_repo: record: %w", err)
	}
	return nil
}

func (r *supabaseListenHistoryRepo) ListByUser(ctx context.Context, userID string) (history []entity.ListenHistory, err error) {
	ctx, span := startRepoSpan(ctx, "listen_history", "ListenHistoryRepository", "ListByUser", "listen_history", "SELECT",
		attribute.String("lookup.kind", "user_id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `
		SELECT id, user_id, COALESCE(track_id::text, ''), stream_id, listened_at, duration_sec
		FROM listen_history
		WHERE user_id = $1
		ORDER BY listened_at DESC`
	rows, queryErr := r.db.Query(ctx, q, userID)
	if queryErr != nil {
		err = fmt.Errorf("listen_history_repo: list: %w", queryErr)
		return nil, err
	}
	defer rows.Close()

	history = []entity.ListenHistory{}
	for rows.Next() {
		var h entity.ListenHistory
		if scanErr := rows.Scan(
			&h.ID, &h.UserID, &h.TrackID, &h.StreamID, &h.ListenedAt, &h.DurationSec,
		); scanErr != nil {
			err = fmt.Errorf("listen_history_repo: scan: %w", scanErr)
			return nil, err
		}
		history = append(history, h)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		err = rowsErr
		return nil, err
	}

	span.SetAttributes(attribute.Int("db.result.row_count", len(history)))
	return history, nil
}
