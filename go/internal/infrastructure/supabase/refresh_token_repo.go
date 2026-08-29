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

// supabaseRefreshTokenRepo implements repository.RefreshTokenRepository using pgx + Supabase PostgreSQL.
type supabaseRefreshTokenRepo struct {
	db pgxDB
}

// NewRefreshTokenRepo returns a RefreshTokenRepository backed by Supabase.
func NewRefreshTokenRepo(db *pgxpool.Pool) repository.RefreshTokenRepository {
	return &supabaseRefreshTokenRepo{db: poolOrNil(db)}
}

const refreshTokenColumns = `id, user_id, token_hash, expires_at, created_at, revoked_at`

// Create persists a newly issued refresh token hash.
func (r *supabaseRefreshTokenRepo) Create(ctx context.Context, token *entity.RefreshToken) (err error) {
	ctx, span := startRepoSpan(ctx, "auth", "RefreshTokenRepository", "Create", "refresh_tokens", "INSERT")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err = r.db.QueryRow(ctx, q, token.UserID, token.TokenHash, token.ExpiresAt).
		Scan(&token.ID, &token.CreatedAt)
	if err != nil {
		return fmt.Errorf("refresh_token_repo: create: %w", err)
	}
	span.SetAttributes(attribute.Bool("db.result.created", true))
	return nil
}

// FindByHash looks up a refresh token by its SHA-256 hash.
// A missing token is not an error — it returns (nil, nil).
func (r *supabaseRefreshTokenRepo) FindByHash(ctx context.Context, tokenHash string) (token *entity.RefreshToken, err error) {
	ctx, span := startRepoSpan(ctx, "auth", "RefreshTokenRepository", "FindByHash", "refresh_tokens", "SELECT",
		attribute.String("lookup.kind", "token_hash"),
	)
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return nil, errDatabaseUnavailable
	}

	const q = `SELECT ` + refreshTokenColumns + ` FROM refresh_tokens WHERE token_hash = $1`

	t := &entity.RefreshToken{}
	err = r.db.QueryRow(ctx, q, tokenHash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.CreatedAt, &t.RevokedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		span.SetAttributes(attribute.Bool("db.result.found", false))
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("refresh_token_repo: find by hash: %w", err)
	}
	span.SetAttributes(attribute.Bool("db.result.found", true))
	return t, nil
}

// Revoke marks a single refresh token as unusable. Revoking an already revoked
// token is a no-op so that a repeated logout stays idempotent.
func (r *supabaseRefreshTokenRepo) Revoke(ctx context.Context, tokenHash string) (err error) {
	ctx, span := startRepoSpan(ctx, "auth", "RefreshTokenRepository", "Revoke", "refresh_tokens", "UPDATE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL`
	tag, execErr := r.db.Exec(ctx, q, tokenHash)
	if execErr != nil {
		return fmt.Errorf("refresh_token_repo: revoke: %w", execErr)
	}
	span.SetAttributes(attribute.Int64("db.result.rows_affected", tag.RowsAffected()))
	return nil
}

// RevokeAllForUser invalidates every active session of a user. Used when the
// password changes and when an account is deleted.
func (r *supabaseRefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID string) (err error) {
	ctx, span := startRepoSpan(ctx, "auth", "RefreshTokenRepository", "RevokeAllForUser", "refresh_tokens", "UPDATE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return errDatabaseUnavailable
	}

	const q = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`
	tag, execErr := r.db.Exec(ctx, q, userID)
	if execErr != nil {
		return fmt.Errorf("refresh_token_repo: revoke all: %w", execErr)
	}
	span.SetAttributes(attribute.Int64("db.result.rows_affected", tag.RowsAffected()))
	return nil
}

// DeleteExpired purges tokens that expired before the given instant. Keeping the
// table bounded is also what enforces the retention window documented for RGPD.
func (r *supabaseRefreshTokenRepo) DeleteExpired(ctx context.Context, before time.Time) (deleted int64, err error) {
	ctx, span := startRepoSpan(ctx, "auth", "RefreshTokenRepository", "DeleteExpired", "refresh_tokens", "DELETE")
	defer finishRepoSpan(span, &err)

	if r.db == nil {
		return 0, errDatabaseUnavailable
	}

	const q = `DELETE FROM refresh_tokens WHERE expires_at < $1`
	tag, execErr := r.db.Exec(ctx, q, before)
	if execErr != nil {
		return 0, fmt.Errorf("refresh_token_repo: delete expired: %w", execErr)
	}
	deleted = tag.RowsAffected()
	span.SetAttributes(attribute.Int64("db.result.rows_affected", deleted))
	return deleted, nil
}
