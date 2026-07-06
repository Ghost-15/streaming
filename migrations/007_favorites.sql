CREATE TABLE IF NOT EXISTS favorites (
    user_id    UUID        NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    track_id   UUID        NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, track_id)
);

CREATE INDEX IF NOT EXISTS idx_favorites_user ON favorites(user_id);
