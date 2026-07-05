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

type supabaseStreamRepo struct {
	db *pgxpool.Pool
}

func NewStreamRepo(db *pgxpool.Pool) repository.StreamRepository {
	return &supabaseStreamRepo{db: db}
}

const streamColumns = `id, title, broadcaster_id, status, started_at, ended_at, listener_count`

func scanStream(row pgx.Row) (*entity.Stream, error) {
	s := &entity.Stream{}
	err := row.Scan(
		&s.ID, &s.Title, &s.BroadcasterID,
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
	return errors.New("not implemented")
}

func (r *supabaseStreamRepo) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error {
	return errors.New("not implemented")
}

func (r *supabaseStreamRepo) IncrementListeners(ctx context.Context, id string, delta int) error {
	return errors.New("not implemented")
}
