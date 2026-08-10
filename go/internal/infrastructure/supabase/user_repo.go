package supabase

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

// supabaseUserRepo implements repository.UserRepository using pgx + Supabase PostgreSQL.
type supabaseUserRepo struct {
	db pgxDB
}

// NewUserRepo returns a UserRepository backed by Supabase.
func NewUserRepo(db *pgxpool.Pool) repository.UserRepository {
	return &supabaseUserRepo{db: poolOrNil(db)}
}

const userColumns = `id, email, password_hash, first_name, last_name, role, created_at, suspended_at`

// FindByEmail looks up a user by email address.
func (r *supabaseUserRepo) FindByEmail(ctx context.Context, email string) (user *entity.User, err error) {
	ctx, span := startRepoSpan(ctx, "auth", "UserRepository", "FindByEmail", "users", "SELECT",
		attribute.String("lookup.kind", "email"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + userColumns + ` FROM users WHERE email = $1`

	u := &entity.User{}
	err = r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &u.SuspendedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		span.SetAttributes(attribute.Bool("db.result.found", false))
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by email: %w", err)
	}
	span.SetAttributes(
		attribute.Bool("db.result.found", true),
		attribute.String("user.role", string(u.Role)),
	)
	return u, nil
}

// FindByID looks up a user by UUID.
func (r *supabaseUserRepo) FindByID(ctx context.Context, id string) (user *entity.User, err error) {
	ctx, span := startRepoSpan(ctx, "auth", "UserRepository", "FindByID", "users", "SELECT",
		attribute.String("lookup.kind", "id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + userColumns + ` FROM users WHERE id = $1`

	u := &entity.User{}
	err = r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt, &u.SuspendedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		span.SetAttributes(attribute.Bool("db.result.found", false))
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo: find by id: %w", err)
	}
	span.SetAttributes(
		attribute.Bool("db.result.found", true),
		attribute.String("user.role", string(u.Role)),
	)
	return u, nil
}

// Create inserts a new user into the database.
func (r *supabaseUserRepo) Create(ctx context.Context, user *entity.User) (err error) {
	ctx, span := startRepoSpan(ctx, "auth", "UserRepository", "Create", "users", "INSERT",
		attribute.String("user.role", string(user.Role)),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		INSERT INTO users (email, password_hash, first_name, last_name, role)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err = r.db.QueryRow(ctx, q,
		user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Role,
	).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return fmt.Errorf("user_repo: create: %w", err)
	}
	span.SetAttributes(attribute.Bool("db.result.created", true))
	return nil
}

// Update updates mutable user fields (role, first_name, last_name).
func (r *supabaseUserRepo) Update(ctx context.Context, user *entity.User) (err error) {
	ctx, span := startRepoSpan(ctx, "auth", "UserRepository", "Update", "users", "UPDATE",
		attribute.String("user.role", string(user.Role)),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `UPDATE users SET role = $1, first_name = $2, last_name = $3 WHERE id = $4`
	tag, execErr := r.db.Exec(ctx, q, user.Role, user.FirstName, user.LastName, user.ID)
	if execErr != nil {
		return fmt.Errorf("user_repo: update: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user_repo: user %s not found", user.ID)
	}
	return nil
}

// Delete removes a user by UUID.
func (r *supabaseUserRepo) Delete(ctx context.Context, id string) (err error) {
	ctx, span := startRepoSpan(ctx, "auth", "UserRepository", "Delete", "users", "DELETE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `DELETE FROM users WHERE id = $1`
	tag, execErr := r.db.Exec(ctx, q, id)
	if execErr != nil {
		return fmt.Errorf("user_repo: delete: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("user_repo: user %s not found", id)
	}
	return nil
}
