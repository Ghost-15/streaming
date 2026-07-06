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

type adminRepo struct {
	db *pgxpool.Pool
}

// NewAdminRepo returns an AdminRepository backed by Supabase.
func NewAdminRepo(db *pgxpool.Pool) repository.AdminRepository {
	return &adminRepo{db: db}
}

// ListUsers returns a paginated list of users and the total count.
func (r *adminRepo) ListUsers(ctx context.Context, page, limit int) (users []entity.User, total int, err error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	ctx, span := startRepoSpan(ctx, "admin", "AdminRepository", "ListUsers", "users", "SELECT",
		attribute.Int("query.page", page),
		attribute.Int("query.limit", limit),
		attribute.Int("query.offset", offset),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, 0, errDatabaseUnavailable
	}

	const countQ = `SELECT COUNT(*) FROM users`
	if err = r.db.QueryRow(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("admin_repo: count users: %w", err)
	}

	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, queryErr := r.db.Query(ctx, q, limit, offset)
	if queryErr != nil {
		err = fmt.Errorf("admin_repo: list users: %w", queryErr)
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var u entity.User
		if scanErr := rows.Scan(
			&u.ID, &u.Email, &u.PasswordHash,
			&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt,
		); scanErr != nil {
			err = fmt.Errorf("admin_repo: scan user: %w", scanErr)
			return nil, 0, err
		}
		u.PasswordHash = ""
		users = append(users, u)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, 0, fmt.Errorf("admin_repo: iterate users: %w", rowsErr)
	}

	span.SetAttributes(
		attribute.Int("db.result.row_count", len(users)),
		attribute.Int("admin.users_total", total),
	)
	return users, total, nil
}

// GetUser looks up a user by ID.
func (r *adminRepo) GetUser(ctx context.Context, id string) (user *entity.User, err error) {
	ctx, span := startRepoSpan(ctx, "admin", "AdminRepository", "GetUser", "users", "SELECT",
		attribute.String("lookup.kind", "id"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `
		SELECT id, email, password_hash, first_name, last_name, role, created_at
		FROM users
		WHERE id = $1`

	u := &entity.User{}
	err = r.db.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.PasswordHash,
		&u.FirstName, &u.LastName, &u.Role, &u.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		span.SetAttributes(attribute.Bool("db.result.found", false))
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("admin_repo: get user: %w", err)
	}
	u.PasswordHash = ""
	span.SetAttributes(
		attribute.Bool("db.result.found", true),
		attribute.String("user.role", string(u.Role)),
	)
	return u, nil
}

// UpdateUserRole changes the role of a user.
func (r *adminRepo) UpdateUserRole(ctx context.Context, id string, role entity.UserRole) (err error) {
	ctx, span := startRepoSpan(ctx, "admin", "AdminRepository", "UpdateUserRole", "users", "UPDATE",
		attribute.String("user.role", string(role)),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `UPDATE users SET role = $1 WHERE id = $2`
	tag, execErr := r.db.Exec(ctx, q, role, id)
	if execErr != nil {
		return fmt.Errorf("admin_repo: update role: %w", execErr)
	}
	rowsAffected := tag.RowsAffected()
	span.SetAttributes(attribute.Int64("db.result.rows_affected", rowsAffected))
	if rowsAffected == 0 {
		return fmt.Errorf("admin_repo: user not found")
	}
	return nil
}

// SuspendUser suspends or reactivates a user account.
func (r *adminRepo) SuspendUser(ctx context.Context, id string, suspend bool) (err error) {
	ctx, span := startRepoSpan(ctx, "admin", "AdminRepository", "SuspendUser", "users", "UPDATE",
		attribute.Bool("user.suspend", suspend),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		UPDATE users
		SET suspended_at = CASE WHEN $2 THEN NOW() ELSE NULL END
		WHERE id = $1`
	tag, execErr := r.db.Exec(ctx, q, id, suspend)
	if execErr != nil {
		return fmt.Errorf("admin_repo: suspend user: %w", execErr)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("admin_repo: user %s not found", id)
	}
	return nil
}

// GetStats returns aggregate user statistics.
func (r *adminRepo) GetStats(ctx context.Context) (stats *entity.AdminStats, err error) {
	ctx, span := startRepoSpan(ctx, "admin", "AdminRepository", "GetStats", "users", "SELECT")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const totalQ = `SELECT COUNT(*) FROM users`
	var total int
	if err = r.db.QueryRow(ctx, totalQ).Scan(&total); err != nil {
		return nil, fmt.Errorf("admin_repo: stats total: %w", err)
	}

	const roleQ = `SELECT role, COUNT(*) FROM users GROUP BY role`
	rows, queryErr := r.db.Query(ctx, roleQ)
	if queryErr != nil {
		return nil, fmt.Errorf("admin_repo: stats by role: %w", queryErr)
	}
	defer rows.Close()

	byRole := make(map[string]int)
	for rows.Next() {
		var role string
		var count int
		if scanErr := rows.Scan(&role, &count); scanErr != nil {
			err = fmt.Errorf("admin_repo: scan role stats: %w", scanErr)
			return nil, err
		}
		byRole[role] = count
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("admin_repo: iterate role stats: %w", rowsErr)
	}

	span.SetAttributes(
		attribute.Int("admin.users_total", total),
		attribute.Int("admin.role_count", len(byRole)),
	)
	return &entity.AdminStats{TotalUsers: total, ByRole: byRole}, nil
}
