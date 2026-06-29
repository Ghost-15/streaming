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

type adminRepo struct {
	db *pgxpool.Pool
}

// NewAdminRepo returns an AdminRepository backed by Supabase.
func NewAdminRepo(db *pgxpool.Pool) repository.AdminRepository {
	return &adminRepo{db: db}
}

// ListUsers returns a paginated list of users and the total count.
func (r *adminRepo) ListUsers(ctx context.Context, page, limit int) ([]entity.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	const countQ = `SELECT COUNT(*) FROM users`
	var total int
	if err := r.db.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin_repo: count users: %w", err)
	}

	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("admin_repo: list users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash,
			&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("admin_repo: scan user: %w", err)
		}
		u.PasswordHash = ""
		users = append(users, u)
	}
	return users, total, nil
}

// GetUser looks up a user by ID.
func (r *adminRepo) GetUser(ctx context.Context, id string) (*entity.User, error) {
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
		return nil, fmt.Errorf("admin_repo: get user: %w", err)
	}
	u.PasswordHash = ""
	return u, nil
}

// UpdateUserRole changes the role of a user.
func (r *adminRepo) UpdateUserRole(ctx context.Context, id string, role entity.UserRole) error {
	const q = `UPDATE users SET role = $1 WHERE id = $2`
	tag, err := r.db.Exec(ctx, q, role, id)
	if err != nil {
		return fmt.Errorf("admin_repo: update role: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("admin_repo: user %s not found", id)
	}
	return nil
}

// SuspendUser suspends (suspended_at = NOW()) or reactivates (suspended_at = NULL) an account.
// The suspended_at column is added by migration 006.
func (r *adminRepo) SuspendUser(ctx context.Context, id string, suspend bool) error {
	const q = `
		UPDATE users
		SET suspended_at = CASE WHEN $2 THEN NOW() ELSE NULL END
		WHERE id = $1`
	tag, err := r.db.Exec(ctx, q, id, suspend)
	if err != nil {
		return fmt.Errorf("admin_repo: suspend user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("admin_repo: user %s not found", id)
	}
	return nil
}

// GetStats returns aggregate user statistics.
func (r *adminRepo) GetStats(ctx context.Context) (*entity.AdminStats, error) {
	const totalQ = `SELECT COUNT(*) FROM users`
	var total int
	if err := r.db.QueryRow(ctx, totalQ).Scan(&total); err != nil {
		return nil, fmt.Errorf("admin_repo: stats total: %w", err)
	}

	const roleQ = `SELECT role, COUNT(*) FROM users GROUP BY role`
	rows, err := r.db.Query(ctx, roleQ)
	if err != nil {
		return nil, fmt.Errorf("admin_repo: stats by role: %w", err)
	}
	defer rows.Close()

	byRole := make(map[string]int)
	for rows.Next() {
		var role string
		var count int
		if err := rows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("admin_repo: scan role stats: %w", err)
		}
		byRole[role] = count
	}

	return &entity.AdminStats{TotalUsers: total, ByRole: byRole}, nil
}
