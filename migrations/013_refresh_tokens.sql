-- Migration 013 — Rotating refresh tokens
-- Access tokens stay short lived (1 h). A refresh token lets a listener keep a
-- session across a long broadcast without re-entering credentials, and gives the
-- server a revocation point that a stateless JWT cannot offer (logout).
-- Only the SHA-256 hash of the opaque value is stored: a dump of this table
-- cannot be replayed against the API.

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT        UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);
