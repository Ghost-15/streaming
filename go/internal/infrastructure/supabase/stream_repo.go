package supabase

import (
	"context"
	"errors"

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

func (r *supabaseStreamRepo) FindByID(ctx context.Context, id string) (stream *entity.Stream, err error) {
	_, span := startRepoSpan(ctx, "streams", "StreamRepository", "FindByID", "streams", "SELECT",
		attribute.String("lookup.kind", "id"),
	)
	defer finishRepoSpan(span, &err)

	return nil, errors.New("not implemented")
}

func (r *supabaseStreamRepo) ListActive(ctx context.Context) (streams []entity.Stream, err error) {
	_, span := startRepoSpan(ctx, "streams", "StreamRepository", "ListActive", "streams", "SELECT",
		attribute.String("stream.status", string(entity.StreamStatusLive)),
	)
	defer finishRepoSpan(span, &err)

	return nil, errors.New("not implemented")
}

func (r *supabaseStreamRepo) Create(ctx context.Context, stream *entity.Stream) (err error) {
	attrs := []attribute.KeyValue{}
	if stream != nil {
		attrs = append(attrs, attribute.String("stream.status", string(stream.Status)))
	}
	_, span := startRepoSpan(ctx, "streams", "StreamRepository", "Create", "streams", "INSERT", attrs...)
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func (r *supabaseStreamRepo) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) (err error) {
	_, span := startRepoSpan(ctx, "streams", "StreamRepository", "UpdateStatus", "streams", "UPDATE",
		attribute.String("stream.status", string(status)),
	)
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}

func (r *supabaseStreamRepo) IncrementListeners(ctx context.Context, id string, delta int) (err error) {
	_, span := startRepoSpan(ctx, "streams", "StreamRepository", "IncrementListeners", "streams", "UPDATE",
		attribute.Int("stream.listener_delta", delta),
	)
	defer finishRepoSpan(span, &err)

	return errors.New("not implemented")
}
