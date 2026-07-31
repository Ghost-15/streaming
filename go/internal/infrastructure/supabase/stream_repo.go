package supabase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

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

func (r *supabaseStreamRepo) FindByID(ctx context.Context, id string) (stream *entity.Stream, err error) {
	ctx, span := startRepoSpan(ctx, "streams", "StreamRepository", "FindByID", "streams", "SELECT",
		attribute.String("lookup.kind", "id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + streamColumns + ` FROM streams WHERE id = $1`
	s, scanErr := scanStream(r.db.QueryRow(ctx, q, id))
	if errors.Is(scanErr, pgx.ErrNoRows) {
		return nil, nil
	}
	if scanErr != nil {
		err = fmt.Errorf("stream_repo: find by id: %w", scanErr)
		return nil, err
	}
	return s, nil
}

func (r *supabaseStreamRepo) ListActive(ctx context.Context) (streams []entity.Stream, err error) {
	ctx, span := startRepoSpan(ctx, "streams", "StreamRepository", "ListActive", "streams", "SELECT",
		attribute.String("stream.status", string(entity.StreamStatusLive)),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + streamColumns + ` FROM streams WHERE status = 'live' ORDER BY started_at DESC`
	rows, queryErr := r.db.Query(ctx, q)
	if queryErr != nil {
		err = fmt.Errorf("stream_repo: list active: %w", queryErr)
		return nil, err
	}
	defer rows.Close()

	streams = []entity.Stream{}
	for rows.Next() {
		s, scanErr := scanStream(rows)
		if scanErr != nil {
			err = fmt.Errorf("stream_repo: scan stream: %w", scanErr)
			return nil, err
		}
		streams = append(streams, *s)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		err = fmt.Errorf("stream_repo: iterate streams: %w", rowsErr)
		return nil, err
	}

	span.SetAttributes(attribute.Int("db.result.row_count", len(streams)))
	return streams, nil
}

func (r *supabaseStreamRepo) Create(ctx context.Context, stream *entity.Stream) (err error) {
	attrs := []attribute.KeyValue{}
	if stream != nil {
		attrs = append(attrs, attribute.String("stream.status", string(stream.Status)))
	}
	ctx, span := startRepoSpan(ctx, "streams", "StreamRepository", "Create", "streams", "INSERT", attrs...)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}
	if stream == nil {
		return errors.New("stream_repo: nil stream")
	}

	const q = `
		INSERT INTO streams (title, broadcaster_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, started_at, ended_at, listener_count`
	err = r.db.QueryRow(ctx, q, stream.Title, stream.BroadcasterID, stream.Status).
		Scan(&stream.ID, &stream.StartedAt, &stream.EndedAt, &stream.ListenerCount)
	if err != nil {
		return fmt.Errorf("stream_repo: create: %w", err)
	}
	return nil
}

func (r *supabaseStreamRepo) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) (err error) {
	ctx, span := startRepoSpan(ctx, "streams", "StreamRepository", "UpdateStatus", "streams", "UPDATE",
		attribute.String("stream.status", string(status)),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	// Compute ended_at in Go to avoid PostgreSQL SQLSTATE 42P08 (inconsistent
	// type inference when the same parameter $2 appears in both SET and CASE).
	var endedAt *time.Time
	if status == entity.StreamStatusEnded {
		now := time.Now()
		endedAt = &now
	}

	const q = `UPDATE streams SET status = $2, ended_at = $3 WHERE id = $1`
	_, execErr := r.db.Exec(ctx, q, id, status, endedAt)
	if execErr != nil {
		return fmt.Errorf("stream_repo: update status: %w", execErr)
	}
	return nil
}

func (r *supabaseStreamRepo) IncrementListeners(ctx context.Context, id string, delta int) (err error) {
	ctx, span := startRepoSpan(ctx, "streams", "StreamRepository", "IncrementListeners", "streams", "UPDATE",
		attribute.Int("stream.listener_delta", delta),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		UPDATE streams
		SET listener_count = GREATEST(listener_count + $2, 0)
		WHERE id = $1`
	tag, execErr := r.db.Exec(ctx, q, id, delta)
	if execErr != nil {
		return fmt.Errorf("stream_repo: increment listeners: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("stream_repo: stream %s not found", id)
	}
	return nil
}
