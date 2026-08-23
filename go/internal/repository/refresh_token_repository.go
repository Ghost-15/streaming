package repository

import (
	"context"
	"time"

	"github.com/Ghost-15/streaming/internal/entity"
)

// RefreshTokenRepository defines the persistence contract for refresh tokens.
// Implemented in internal/infrastructure/supabase/refresh_token_repo.go
// Never import handler or usecase packages here — dependency rule.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
