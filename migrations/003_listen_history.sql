-- Migration 003 — Listen history table
-- Used by the recommendation engine (US-025) and analytics.

CREATE TABLE IF NOT EXISTS listen_history (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id     UUID NOT NULL,
    stream_id    UUID REFERENCES streams(id) ON DELETE SET NULL,
    listened_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    duration_sec INT NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, track_id, listened_at)
);

CREATE INDEX idx_listen_history_user ON listen_history(user_id);
CREATE INDEX idx_listen_history_track ON listen_history(track_id);

-- RLS
ALTER TABLE listen_history ENABLE ROW LEVEL SECURITY;

-- Users can insert their own history
CREATE POLICY listen_history_insert_own ON listen_history
    FOR INSERT WITH CHECK (user_id = auth.uid());

-- Users can read their own history
CREATE POLICY listen_history_select_own ON listen_history
    FOR SELECT USING (user_id = auth.uid());

-- Admins can read everything (analytics)
CREATE POLICY listen_history_admin_all ON listen_history
    FOR ALL USING (
        EXISTS (SELECT 1 FROM users WHERE id = auth.uid() AND role = 'admin')
    );
