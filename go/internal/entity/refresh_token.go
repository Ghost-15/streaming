package entity

import "time"

// RefreshToken is a persisted credential exchanged for a new access token.
// The opaque value is never stored — only its SHA-256 hash — so a database
// leak cannot be replayed against the API.
type RefreshToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}

// IsUsable reports whether the token can still be exchanged for an access token.
func (t *RefreshToken) IsUsable(now time.Time) bool {
	return !t.IsRevoked() && !t.IsExpired(now)
}
