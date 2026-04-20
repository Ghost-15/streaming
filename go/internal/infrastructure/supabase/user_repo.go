package supabase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// supabaseUserRepo implements repository.UserRepository using pgx + Supabase PostgreSQL.
// OTEL spans will be added in Sprint 2 (US-008).
type supabaseUserRepo struct {
	db *pgxpool.Pool
}

// NewUserRepo returns a UserRepository backed by Supabase.
func NewUserRepo(db *pgxpool.Pool) repository.UserRepository {
	return &supabaseUserRepo{db: db}
}

func (r *supabaseUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	// TODO Sprint 1 — US-001: SELECT * FROM users WHERE email = $1
	return nil, errors.New("not implemented")
}

func (r *supabaseUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	// TODO Sprint 1 — US-001: SELECT * FROM users WHERE id = $1
	return nil, errors.New("not implemented")
}

func (r *supabaseUserRepo) Create(ctx context.Context, user *entity.User) error {
	// TODO Sprint 1 — US-001: INSERT INTO users ...
	return errors.New("not implemented")
}

func (r *supabaseUserRepo) Update(ctx context.Context, user *entity.User) error {
	// TODO Sprint 3 — US-013: UPDATE users SET ...
	return errors.New("not implemented")
}

func (r *supabaseUserRepo) Delete(ctx context.Context, id string) error {
	// TODO Sprint 3 — US-013: DELETE FROM users WHERE id = $1
	return errors.New("not implemented")
}
