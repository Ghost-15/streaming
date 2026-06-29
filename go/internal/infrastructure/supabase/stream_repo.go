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

// supabaseStreamRepo implements repository.StreamRepository.
type supabaseStreamRepo struct {
	db *pgxpool.Pool
}

// NewStreamRepo returns a StreamRepository backed by Supabase.
func NewStreamRepo(db *pgxpool.Pool) repository.StreamRepository {
	return &supabaseStreamRepo{db: db}
}

const streamColumns = `id, title, broadcaster_id, stream_url, status, started_at, ended_at, listener_count`

func scanStream(row pgx.Row) (*entity.Stream, error) {
	s := &entity.Stream{}
	err := row.Scan(
		&s.ID, &s.Title, &s.BroadcasterID, &s.StreamURL,
		&s.Status, &s.StartedAt, &s.EndedAt, &s.ListenerCount,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *supabaseStreamRepo) FindByID(ctx context.Context, id string) (*entity.Stream, error) {
	const q = `SELECT ` + streamColumns + ` FROM streams WHERE id = $1`
	s, err := scanStream(r.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("stream_repo: find by id: %w", err)
	}
	return s, nil
}

func (r *supabaseStreamRepo) ListActive(ctx context.Context) ([]entity.Stream, error) {
	const q = `SELECT ` + streamColumns + ` FROM streams WHERE status = 'live' ORDER BY started_at DESC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("stream_repo: list active: %w", err)
	}
	defer rows.Close()

	streams := []entity.Stream{}
	for rows.Next() {
		s, err := scanStream(rows)
		if err != nil {
			return nil, fmt.Errorf("stream_repo: scan stream: %w", err)
		}
		streams = append(streams, *s)
	}
	return streams, rows.Err()
}

func (r *supabaseStreamRepo) Create(ctx context.Context, stream *entity.Stream) error {
	const q = `
		INSERT INTO streams (title, broadcaster_id, stream_url, status)
		VALUES ($1, $2, $3, 'live')
		RETURNING id, started_at, status, listener_count`
	err := r.db.QueryRow(ctx, q, stream.Title, stream.BroadcasterID, stream.StreamURL).
		Scan(&stream.ID, &stream.StartedAt, &stream.Status, &stream.ListenerCount)
	if err != nil {
		return fmt.Errorf("stream_repo: create: %w", err)
	}
	return nil
}

func (r *supabaseStreamRepo) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error {
	// When a stream ends, stamp ended_at; otherwise leave it NULL.
	const q = `
		UPDATE streams
		SET status = $1,
		    ended_at = CASE WHEN $1 = 'ended' THEN NOW() ELSE ended_at END
		WHERE id = $2`
	tag, err := r.db.Exec(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("stream_repo: update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("stream_repo: stream %s not found", id)
	}
	return nil
}

func (r *supabaseStreamRepo) IncrementListeners(ctx context.Context, id string, delta int) error {
	const q = `
		UPDATE streams
		SET listener_count = GREATEST(listener_count + $2, 0)
		WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, delta)
	if err != nil {
		return fmt.Errorf("stream_repo: increment listeners: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("stream_repo: stream %s not found", id)
	}
	return nil
}
