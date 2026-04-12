package supabase

import (
	"context"
	"errors"

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

func (r *supabaseStreamRepo) FindByID(ctx context.Context, id string) (*entity.Stream, error) {
	// TODO Sprint 1 — US-003
	return nil, errors.New("not implemented")
}

func (r *supabaseStreamRepo) ListActive(ctx context.Context) ([]entity.Stream, error) {
	// TODO Sprint 1 — US-003: SELECT * FROM streams WHERE status = 'live'
	return nil, errors.New("not implemented")
}

func (r *supabaseStreamRepo) Create(ctx context.Context, stream *entity.Stream) error {
	// TODO Sprint 1 — US-003
	return errors.New("not implemented")
}

func (r *supabaseStreamRepo) UpdateStatus(ctx context.Context, id string, status entity.StreamStatus) error {
	// TODO Sprint 1 — US-003
	return errors.New("not implemented")
}

func (r *supabaseStreamRepo) IncrementListeners(ctx context.Context, id string, delta int) error {
	// TODO Sprint 1 — US-003: UPDATE streams SET listener_count = listener_count + $2
	return errors.New("not implemented")
}
