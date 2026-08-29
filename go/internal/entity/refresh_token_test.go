package entity_test

import (
	"testing"
	"time"

	"github.com/Ghost-15/streaming/internal/entity"
)

func TestRefreshToken_IsRevoked(t *testing.T) {
	active := &entity.RefreshToken{}
	if active.IsRevoked() {
		t.Error("IsRevoked() = true on a token that was never revoked")
	}

	now := time.Now()
	revoked := &entity.RefreshToken{RevokedAt: &now}
	if !revoked.IsRevoked() {
		t.Error("IsRevoked() = false on a revoked token")
	}
}

func TestRefreshToken_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		want      bool
	}{
		{name: "future expiry", expiresAt: now.Add(time.Hour), want: false},
		{name: "past expiry", expiresAt: now.Add(-time.Hour), want: true},
		{name: "expiring exactly now", expiresAt: now, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := &entity.RefreshToken{ExpiresAt: tt.expiresAt}
			if got := token.IsExpired(now); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefreshToken_IsUsable(t *testing.T) {
	now := time.Now()
	revokedAt := now.Add(-time.Minute)

	tests := []struct {
		name  string
		token *entity.RefreshToken
		want  bool
	}{
		{
			name:  "active and unexpired",
			token: &entity.RefreshToken{ExpiresAt: now.Add(time.Hour)},
			want:  true,
		},
		{
			name:  "revoked",
			token: &entity.RefreshToken{ExpiresAt: now.Add(time.Hour), RevokedAt: &revokedAt},
			want:  false,
		},
		{
			name:  "expired",
			token: &entity.RefreshToken{ExpiresAt: now.Add(-time.Hour)},
			want:  false,
		},
		{
			name:  "revoked and expired",
			token: &entity.RefreshToken{ExpiresAt: now.Add(-time.Hour), RevokedAt: &revokedAt},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsUsable(now); got != tt.want {
				t.Errorf("IsUsable() = %v, want %v", got, tt.want)
			}
		})
	}
}
