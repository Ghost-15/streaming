package mock

import (
	"context"
	"time"

	"github.com/Ghost-15/streaming/internal/entity"
	"github.com/Ghost-15/streaming/internal/repository"
)

var _ repository.RefreshTokenRepository = (*MockRefreshTokenRepository)(nil)

// MockRefreshTokenRepository records issued tokens in memory. Unset function
// fields fall back to an in-memory store, so a test only overrides what it needs.
type MockRefreshTokenRepository struct {
	CreateFn           func(ctx context.Context, token *entity.RefreshToken) error
	FindByHashFn       func(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	RevokeFn           func(ctx context.Context, tokenHash string) error
	RevokeAllForUserFn func(ctx context.Context, userID string) error
	DeleteExpiredFn    func(ctx context.Context, before time.Time) (int64, error)

	Stored []*entity.RefreshToken
}

func (m *MockRefreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, token)
	}
	token.ID = "refresh-" + token.TokenHash[:8]
	token.CreatedAt = time.Now()
	m.Stored = append(m.Stored, token)
	return nil
}

func (m *MockRefreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	if m.FindByHashFn != nil {
		return m.FindByHashFn(ctx, tokenHash)
	}
	for _, t := range m.Stored {
		if t.TokenHash == tokenHash {
			return t, nil
		}
	}
	return nil, nil
}

func (m *MockRefreshTokenRepository) Revoke(ctx context.Context, tokenHash string) error {
	if m.RevokeFn != nil {
		return m.RevokeFn(ctx, tokenHash)
	}
	now := time.Now()
	for _, t := range m.Stored {
		if t.TokenHash == tokenHash && t.RevokedAt == nil {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	if m.RevokeAllForUserFn != nil {
		return m.RevokeAllForUserFn(ctx, userID)
	}
	now := time.Now()
	for _, t := range m.Stored {
		if t.UserID == userID && t.RevokedAt == nil {
			t.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) (int64, error) {
	if m.DeleteExpiredFn != nil {
		return m.DeleteExpiredFn(ctx, before)
	}
	kept := m.Stored[:0]
	var deleted int64
	for _, t := range m.Stored {
		if t.ExpiresAt.Before(before) {
			deleted++
			continue
		}
		kept = append(kept, t)
	}
	m.Stored = kept
	return deleted, nil
}
