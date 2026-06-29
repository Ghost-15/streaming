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

// supabaseUserRepo implements repository.UserRepository using pgx + Supabase PostgreSQL.
// OTEL spans will be added in Sprint 2 (US-008).
type supabaseUserRepo struct {
	db *pgxpool.Pool
}

// NewUserRepo returns a UserRepository backed by Supabase.
func NewUserRepo(db *pgxpool.Pool) repository.UserRepository {
	return &supabaseUserRepo{db: db}
}

// FindByEmail looks up a user by email address.
// Returns nil, nil if no user exists with that email (not found ≠ error).
func (r *supabaseUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at
		FROM users
		WHERE email = $1`

	u := &entity.User{}
	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by email: %w", err)
	}
	return u, nil
}

// FindByID looks up a user by UUID.
// Returns nil, nil if no user exists with that ID.
func (r *supabaseUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at
		FROM users
		WHERE id = $1`

	u := &entity.User{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by id: %w", err)
	}
	return u, nil
}

// Create inserts a new user into the database.
// The database generates the UUID and created_at; both are written back into user.
func (r *supabaseUserRepo) Create(ctx context.Context, user *entity.User) error {
	const q = `
		INSERT INTO users (email, password_hash, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q,
		user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Role,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("user_repo: create: %w", err)
	}
	return nil
}

// Update updates mutable user fields (role, first_name, last_name).
// Sprint 3 — US-013.
func (r *supabaseUserRepo) Update(ctx context.Context, user *entity.User) error {
	const q = `UPDATE users SET role = $1, first_name = $2, last_name = $3 WHERE id = $4`
	tag, err := r.db.Exec(ctx, q, user.Role, user.FirstName, user.LastName, user.ID)
	if err != nil {
		return fmt.Errorf("user_repo: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user_repo: user %s not found", user.ID)
	}
	return nil
}

// Delete removes a user by UUID.
// Sprint 3 — US-013.
func (r *supabaseUserRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("user_repo: delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user_repo: user %s not found", id)
	}
	return nil
}
